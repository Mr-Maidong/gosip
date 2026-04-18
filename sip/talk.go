package sipapi

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	sdp "github.com/panjjo/gosdp"
	"github.com/panjjo/gosip/db"
	"github.com/panjjo/gosip/m"
	sip "github.com/panjjo/gosip/sip/s"
	"github.com/panjjo/gosip/utils"
	"github.com/sirupsen/logrus"
)

// sip 请求对讲
func SipTalk(data *Streams) (*Streams, error) {

	channel := Channels{ChannelID: data.ChannelID}
	if err := db.Get(db.DBClient, &channel); err != nil {
		if db.RecordNotFound(err) {
			return nil, errors.New("通道不存在")
		}
		return nil, err
	}

	// 若通道的播放类型为空，默认按 push 处理
	if channel.StreamType == "" {
		channel.StreamType = m.StreamTypePush
	}

	data.DeviceID = channel.DeviceID
	data.StreamType = channel.StreamType
	// 使用通道的播放模式进行处理
	switch channel.StreamType {
	default:
		user, ok := _activeDevices.Get(channel.DeviceID)
		if !ok {
			return nil, errors.New("设备已离线")
		}
		// GB28181推流
		ssrcLock.Lock()
		data.ssrc = getSSRC(data.T)
		data.StreamID = generateTalkStreamID(data.DeviceID, data.ChannelID)

		// 在 ZLM 中开启 RTP 服务器，指定自定义 streamId（对讲使用被动模式）
		rtpReq := zlmStartSendRtpPassivReq{
			Vhost:  "__defaultVhost__", // 默认虚拟主机
			App:    "rtp",              // App 名称
			Stream: data.StreamID,      // 格式: talk_channelId
			Ssrc:   "0",                // ssrc
		}

		rtpResp, err := zlmStartSendRtpPassive(rtpReq)
		logrus.Infoln("开启 ZLM RTP 服务器（对讲）", rtpReq, rtpResp)
		if err != nil {
			ssrcLock.Unlock()
			return nil, fmt.Errorf("开启 ZLM RTP 服务器失败: %v", err)
		}

		// 更新媒体服务器端口（如果 ZLM 返回了新端口）
		if rtpResp.Port > 0 {
			_sysinfo.MediaServerRtpPort = rtpResp.Port
		}

		// 成功后保存
		db.Create(db.DBClient, data)
		ssrcLock.Unlock()

		data, err = sipTalkPush(data, channel, user)
		if err != nil {
			return nil, fmt.Errorf("获取视频失败:%v", err)
		}
	}

	data.HTTP = fmt.Sprintf("%s/rtp/%s/hls.m3u8", config.Media.HTTP, data.StreamID)
	data.RTMP = fmt.Sprintf("%s/rtp/%s", config.Media.RTMP, data.StreamID)
	data.RTSP = fmt.Sprintf("%s/rtp/%s", config.Media.RTSP, data.StreamID)
	data.WSFLV = fmt.Sprintf("%s/rtp/%s.live.flv", config.Media.WS, data.StreamID)

	data.Ext = time.Now().Unix() + 2*60 // 2分钟等待时间
	StreamList.Response.Store(data.StreamID, data)
	if data.T == 0 {
		StreamList.Succ.Store(data.ChannelID, data)
	}
	db.Save(db.DBClient, data)
	return data, nil
}

func sipTalkPush(data *Streams, channel Channels, device Devices) (*Streams, error) {
	var (
		s sdp.Session
		b []byte
	)
	name := "Talk"
	// 根据设备传输协议设置 SDP 协议
	protocal := "RTP/AVP"
	if strings.ToLower(device.TransPort) == "tcp" {
		protocal = "TCP/RTP/AVP"
	}

	// 音频媒体描述
	audio := sdp.Media{
		Description: sdp.MediaDescription{
			Type:    "audio",
			Port:    _sysinfo.MediaServerRtpPort,
			Formats: []string{"8"},
			// Formats:  []string{"8 104"},
			Protocol: protocal,
		},
	}
	audio.AddAttribute("sendonly")
	audio.AddAttribute("rtpmap", "8", "PCMA/8000")
	// audio.AddAttribute("rtpmap", "104", "mpeg4-generic/32000")

	// 确定收流地址：优先使用设备指定的 StreamIP，否则使用媒体服务器默认 IP
	streamIP := _sysinfo.MediaServerRtpIP
	if device.StreamIP != "" {
		if ip := net.ParseIP(device.StreamIP); ip != nil {
			streamIP = ip
		}
	}

	// defining message
	msg := &sdp.Message{
		Origin: sdp.Origin{
			Username: _serverDevices.DeviceID, // 媒体服务器id
			Address:  streamIP.String(),
		},
		Name: name,
		Connection: sdp.ConnectionData{
			IP:  streamIP,
			TTL: 0,
		},
		Timing: []sdp.Timing{
			{
				Start: data.S,
				End:   data.E,
			},
		},
		Medias: []sdp.Media{audio}, // 同时包含视频和音频
		SSRC:   data.ssrc,
	}

	// appending message to session
	s = msg.Append(s)
	// appending session to byte buffer
	b = s.AppendTo(b)
	uri, _ := sip.ParseURI(channel.URIStr)
	channel.addr = &sip.Address{URI: uri}
	_serverDevices.addr.Params.Add("tag", sip.String{Str: utils.RandString(20)})

	// 根据设备传输协议设置 Via 的 Transport
	transport := "UDP"
	if strings.ToLower(device.TransPort) == "tcp" {
		transport = "TCP"
	}

	hb := sip.NewHeaderBuilder().SetTo(channel.addr).SetFrom(_serverDevices.addr).AddVia(&sip.ViaHop{
		Transport: transport,
		Params:    sip.NewParams().Add("branch", sip.String{Str: sip.GenerateBranch()}),
	}).SetContentType(&sip.ContentTypeSDP).SetMethod(sip.INVITE).SetContact(_serverDevices.addr)
	req := sip.NewRequest("", sip.INVITE, channel.addr.URI, sip.DefaultSipVersion, hb.Build(), b)
	req.SetDestination(device.source)
	req.AppendHeader(&sip.GenericHeader{HeaderName: "Subject", Contents: fmt.Sprintf("%s:%s,%s:%s", channel.ChannelID, data.StreamID, _serverDevices.DeviceID, data.StreamID)})
	req.SetRecipient(channel.addr.URI)
	// 根据设备的传输方式发送请求
	var tx *sip.Transaction
	var err error
	if strings.ToLower(device.TransPort) == "tcp" {
		tx, err = srv.RequestWithProtocol(req, "tcp")
	} else {
		tx, err = srv.Request(req) // 默认UDP
	}
	if err != nil {
		logrus.Warningln("sipTalkPush fail.id:", device.DeviceID, channel.ChannelID, "err:", err)
		return data, err
	}
	// response
	response, err := sipResponse(tx)
	if err != nil {
		logrus.Warningln("sipTalkPush response fail.id:", device.DeviceID, channel.ChannelID, "err:", err)
		return data, err
	}
	data.Resp = response
	// ACK
	tx.Request(sip.NewRequestFromResponse(sip.ACK, response))

	callid, _ := response.CallID()
	data.CallID = string(*callid)

	cseq, _ := response.CSeq()
	if cseq != nil {
		data.CseqNo = cseq.SeqNo
	}

	from, _ := response.From()
	to, _ := response.To()
	for k, v := range to.Params.Items() {
		data.Ttag[k] = v.String()
	}
	for k, v := range from.Params.Items() {
		data.Ftag[k] = v.String()
	}
	data.Status = 0

	return data, err
}

// sip 停止对讲
func SipStopTalk(ssrc string) {
	zlmCloseStream(ssrc)
	// 关闭 ZLM RTP 服务器
	if err := zlmCloseRtpServer(ssrc); err != nil {
		logrus.Errorln("关闭 ZLM RTP 服务器失败:", err)
	}

	data, ok := StreamList.Response.Load(ssrc)
	if !ok {
		return
	}
	talk := data.(*Streams)
	logrus.Infoln("SipStopTalk", talk.StreamType, m.StreamTypePush)
	if talk.StreamType == m.StreamTypePush {
		// 推流，需要发送关闭请求
		resp := talk.Resp
		u, ok := _activeDevices.Load(talk.DeviceID)
		if !ok {
			// 设备已离线，直接清理
			logrus.Debugln("[SipStopTalk] device offline, skip BYE:", talk.DeviceID)
			talk.Status = 1
			talk.Stop = true
			talk.Msg = "设备已离线"
			db.Save(db.DBClient, talk)
			StreamList.Response.Delete(ssrc)
			if talk.T == 0 {
				StreamList.Succ.Delete(talk.ChannelID)
			}
			return
		}
		user := u.(Devices)
		req := sip.NewRequestFromResponse(sip.BYE, resp)
		req.SetDestination(user.source)
		// 根据设备的传输方式发送请求
		var tx *sip.Transaction
		var err error
		if strings.ToLower(user.TransPort) == "tcp" {
			tx, err = srv.RequestWithProtocol(req, "tcp")
		} else {
			tx, err = srv.Request(req) // 默认UDP
		}
		if err != nil {
			// 发送失败（如连接断开），直接清理
			logrus.Warningln("[SipStopTalk] send BYE failed:", talk.DeviceID, talk.ChannelID, "err:", err)
			talk.Status = 1
			talk.Stop = true
			talk.Msg = "发送 BYE 失败: " + err.Error()
			db.Save(db.DBClient, talk)
			StreamList.Response.Delete(ssrc)
			if talk.T == 0 {
				StreamList.Succ.Delete(talk.ChannelID)
			}
			return
		}

		// 等待响应
		_, err = sipResponse(tx)
		if err != nil {
			logrus.Warnln("[SipStopTalk] BYE response failed:", talk.DeviceID, talk.ChannelID, "err:", err)
			talk.Msg = err.Error()
		} else {
			talk.Status = 1
			talk.Stop = true
		}
		db.Save(db.DBClient, talk)
	}
	StreamList.Response.Delete(ssrc)
	if talk.T == 0 {
		StreamList.Succ.Delete(talk.ChannelID)
	}
}
