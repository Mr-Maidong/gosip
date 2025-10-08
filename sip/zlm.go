package sipapi

import (
	"fmt"
	"net/url"

	"github.com/panjjo/gosip/utils"
	"github.com/sirupsen/logrus"
)

type zlmGetMediaListReq struct {
	vhost    string
	schema   string
	streamID string
	app      string
}
type zlmGetMediaListResp struct {
	Code int                       `json:"code"`
	Data []zlmGetMediaListDataResp `json:"data"`
}
type zlmGetMediaListDataResp struct {
	App        string                  `json:"app"`
	Stream     string                  `json:"stream"`
	Schema     string                  `json:"schema"`
	OriginType int                     `json:"originType"`
	Tracks     []zlmGetMediaListTracks `json:"tracks"`
}
type zlmGetMediaListTracks struct {
	Type    int `json:"codec_type"`
	CodecID int `json:"codec_id"`
	Height  int `json:"height"`
	Width   int `json:"width"`
	FPS     int `json:"fps"`
}

// zlm 获取流列表信息
func zlmGetMediaList(req zlmGetMediaListReq) zlmGetMediaListResp {
	res := zlmGetMediaListResp{}
	reqStr := "/index/api/getMediaList?secret=" + config.Media.Secret
	if req.streamID != "" {
		reqStr += "&stream=" + req.streamID
	}
	if req.app != "" {
		reqStr += "&app=" + req.app
	}
	if req.schema != "" {
		reqStr += "&schema=" + req.schema
	}
	if req.vhost != "" {
		reqStr += "&vhost=" + req.vhost
	}
	body, err := utils.GetRequest(config.Media.RESTFUL + reqStr)
	if err != nil {
		logrus.Errorln("get stream mediaList fail,", err)
		return res
	}
	if err = utils.JSONDecode(body, &res); err != nil {
		logrus.Errorln("get stream mediaList fail,", err)
		return res
	}
	logrus.Traceln("zlmGetMediaList ", string(body), req.streamID)
	return res
}

var zlmDeviceVFMap = map[int]string{
	0: "H264",
	1: "H265",
	2: "ACC",
	3: "G711A",
	4: "G711U",
}

func transZLMDeviceVF(t int) string {
	if v, ok := zlmDeviceVFMap[t]; ok {
		return v
	}
	return "undefind"
}

type rtpInfo struct {
	Code  int  `json:"code"`
	Exist bool `json:"exist"`
}

// 获取流在zlm上的信息
func zlmGetMediaInfo(ssrc string) rtpInfo {
	res := rtpInfo{}
	body, err := utils.GetRequest(config.Media.RESTFUL + "/index/api/getRtpInfo?secret=" + config.Media.Secret + "&stream_id=" + ssrc)
	if err != nil {
		logrus.Errorln("get stream rtpInfo fail,", err)
		return res
	}
	if err = utils.JSONDecode(body, &res); err != nil {
		logrus.Errorln("get stream rtpInfo fail,", err)
		return res
	}
	return res
}

// zlm 关闭流
func zlmCloseStream(ssrc string) {
	utils.GetRequest(config.Media.RESTFUL + "/index/api/close_streams?secret=" + config.Media.Secret + "&stream=" + ssrc)
}

// zlm 开始录制视频流
func zlmStartRecord(values url.Values) error {
	body, err := utils.GetRequest(config.Media.RESTFUL + "/index/api/startRecord?" + values.Encode())
	if err != nil {
		return err
	}
	tmp := map[string]interface{}{}
	err = utils.JSONDecode(body, &tmp)
	if err != nil {
		return err
	}
	if code, ok := tmp["code"]; !ok || fmt.Sprint(code) != "0" {
		return utils.NewError(nil, tmp)
	}
	return nil
}

// zlm 停止录制
func zlmStopRecord(values url.Values) error {
	body, err := utils.GetRequest(config.Media.RESTFUL + "/index/api/stopRecord?" + values.Encode())
	if err != nil {
		return err
	}
	tmp := map[string]interface{}{}
	err = utils.JSONDecode(body, &tmp)
	if err != nil {
		return err
	}
	if code, ok := tmp["code"]; !ok || fmt.Sprint(code) != "0" {
		return utils.NewError(nil, tmp)
	}
	return nil
}

// ZLM openRtpServer 请求结构
type zlmOpenRtpServerReq struct {
	Port      string `json:"port"`
	StreamID  string `json:"stream_id"`
	EnableTCP string `json:"enable_tcp"`
}

// ZLM openRtpServer 响应结构
type zlmOpenRtpServerResp struct {
	Code int `json:"code"`
	Port int `json:"port"`
}

// zlm 开启 RTP 服务器，指定 streamId
func zlmOpenRtpServer(req zlmOpenRtpServerReq) (zlmOpenRtpServerResp, error) {
	res := zlmOpenRtpServerResp{}

	params := url.Values{}
	params.Set("secret", config.Media.Secret)
	params.Set("stream_id", req.StreamID)
	params.Set("enable_tcp", req.EnableTCP)
	params.Set("port", req.Port)
	if req.StreamID != "" {
		params.Set("stream_id", req.StreamID)
	}

	body, err := utils.GetRequest(config.Media.RESTFUL + "/index/api/openRtpServer?" + params.Encode())
	if err != nil {
		logrus.Errorln("zlm openRtpServer fail,", err)
		return res, err
	}

	if err = utils.JSONDecode(body, &res); err != nil {
		logrus.Errorln("zlm openRtpServer decode fail,", err)
		return res, err
	}

	logrus.Traceln("zlmOpenRtpServer success", string(body), req.StreamID)
	return res, nil
}

// zlm 关闭 RTP 服务器
func zlmCloseRtpServer(streamID string) error {
	params := url.Values{}
	params.Set("secret", config.Media.Secret)
	params.Set("stream_id", streamID)

	body, err := utils.GetRequest(config.Media.RESTFUL + "/index/api/closeRtpServer?" + params.Encode())
	if err != nil {
		logrus.Errorln("zlm closeRtpServer fail,", err)
		return err
	}

	tmp := map[string]interface{}{}
	if err = utils.JSONDecode(body, &tmp); err != nil {
		logrus.Errorln("zlm closeRtpServer decode fail,", err)
		return err
	}

	if code, ok := tmp["code"]; !ok || fmt.Sprint(code) != "0" {
		return utils.NewError(nil, tmp)
	}

	logrus.Traceln("zlmCloseRtpServer success", streamID)
	return nil
}
