package sipapi

import (
	"encoding/xml"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/panjjo/gosip/db"
	"github.com/panjjo/gosip/m"
	sip "github.com/panjjo/gosip/sip/s"
	"github.com/panjjo/gosip/utils"
	"github.com/sirupsen/logrus"
)

var (
	// sip服务用户信息
	_serverDevices Devices
	srv            *sip.Server
)

// logActiveDeviceStore 记录设备缓存存入日志
func logActiveDeviceStore(action string, device *Devices) {
	logrus.Infof("[ActiveDevice] Store: action=%s, deviceID=%s, Regist=%v, Online=%v, StreamIP=%s, SipIP=%s",
		action, device.DeviceID, device.Regist, device.Online, device.StreamIP, device.SipIP)
}

// logActiveDeviceDelete 记录设备缓存删除日志
func logActiveDeviceDelete(action string, deviceID string) {
	logrus.Infof("[ActiveDevice] Delete: action=%s, deviceID=%s", action, deviceID)
}

// Devices NVR  设备信息
type Devices struct {
	db.DBModel
	// Name 设备名称
	Name string `json:"name" gorm:"column:name" `
	// DeviceID 设备id
	DeviceID string `json:"deviceid" gorm:"column:deviceid"`
	// Region 设备域
	Region string `json:"region" gorm:"column:region"`
	// Host Via 地址
	Host string `json:"host" gorm:"column:host"`
	// Port via 端口
	Port string `json:"port" gorm:"column:port"`
	// TransPort via transport
	TransPort string `json:"transport" gorm:"column:transport"`
	// Proto 协议
	Proto string `json:"proto" gorm:"column:proto"`
	// Rport via rport
	Rport string `json:"report" gorm:"column:report"`
	// RAddr via recevied
	RAddr string `json:"raddr"  gorm:"column:raddr"`
	// Manufacturer 制造厂商
	Manufacturer string `xml:"Manufacturer"  json:"manufacturer"  gorm:"column:manufacturer"`
	// 设备类型DVR，NVR
	DeviceType string `xml:"DeviceType"  json:"devicetype"  gorm:"column:devicetype"`
	// Firmware 固件版本
	Firmware string ` json:"firmware"  gorm:"column:firmware"`
	// Model 型号
	Model  string `json:"model"  gorm:"column:model"`
	URIStr string `json:"uri"  gorm:"column:uri"`
	// ActiveAt 最后心跳检测时间
	ActiveAt int64 `json:"active" gorm:"column:active"`
	// Regist 是否注册
	Regist bool `json:"regist" gorm:"column:regist"`
	// Online 是否在线
	Online bool `json:"online" gorm:"column:online"`
	// PWD 密码
	PWD string `json:"pwd" gorm:"column:pwd"`
	// Source 设备发送地址（字符串）
	Source string `json:"source" gorm:"column:source"`
	// SipIP SIP 信令通讯 IP（优先使用）
	SipIP string `json:"sipip" gorm:"column:sipip"`
	// StreamIP 设备接收媒体流 IP（优先使用）
	StreamIP string `json:"streamip" gorm:"column:streamip"`
	// Longitude 经度
	Longitude float64 `json:"longitude" gorm:"column:longitude;type:decimal(10,6)"`
	// Latitude 纬度
	Latitude float64 `json:"latitude" gorm:"column:latitude;type:decimal(10,6)"`
	// GpsTime GPS时间
	GpsTime time.Time `json:"gps_time" gorm:"column:gps_time;type:datetime"`
	// Subscribe 订阅设置 (JSON格式)
	// 格式：{"position":true, "alarm":false}
	Subscribe db.M `json:"subscribe" gorm:"column:subscribe;type:json"`

	Sys m.SysInfo `json:"sysinfo" gorm:"-"`

	//----
	addr   *sip.Address `gorm:"-"`
	source net.Addr     `gorm:"-"`
}

// Channels 摄像头通道信息
type Channels struct {
	db.DBModel
	// ChannelID 通道编码
	ChannelID string `xml:"DeviceID" json:"channelid" gorm:"column:channelid"`
	// DeviceID 设备编号
	DeviceID string `xml:"-" json:"deviceid"  gorm:"column:deviceid"`
	// Memo 备注（用来标示通道信息）
	MeMo string `json:"memo"  gorm:"column:memo"`
	// Name 通道名称（设备端设置名称）
	Name         string `xml:"Name" json:"name"  gorm:"column:name"`
	Manufacturer string `xml:"Manufacturer" json:"manufacturer"  gorm:"column:manufacturer"`
	Model        string `xml:"Model" json:"model"  gorm:"column:model"`
	Owner        string `xml:"Owner"  json:"owner"  gorm:"column:owner"`
	CivilCode    string `xml:"CivilCode" json:"civilcode"  gorm:"column:civilcode"`
	// Address ip地址
	Address     string `xml:"Address"  json:"address"  gorm:"column:address"`
	Parental    int    `xml:"Parental"  json:"parental"  gorm:"column:parental"`
	SafetyWay   int    `xml:"SafetyWay"  json:"safetyway"  gorm:"column:safetyway"`
	RegisterWay int    `xml:"RegisterWay"  json:"registerway"  gorm:"column:registerway"`
	Secrecy     int    `xml:"Secrecy" json:"secrecy"  gorm:"column:secrecy"`
	// Status 状态  on 在线
	Status string `xml:"Status"  json:"status"  gorm:"column:status"`
	// Active 最后活跃时间
	Active int64  `json:"active"  gorm:"column:active"`
	URIStr string ` json:"uri"  gorm:"column:uri"`

	// 视频编码格式
	VF string ` json:"vf"  gorm:"column:vf"`
	// 视频高
	Height int `json:"height"  gorm:"column:height"`
	// 视频宽
	Width int `json:"width"  gorm:"column:width"`
	// 视频FPS
	FPS int `json:"fps"  gorm:"column:fps"`
	//  pull 媒体服务器主动拉流，push 监控设备主动推流
	StreamType string `json:"streamtype" gorm:"column:streamtype;default:'push'"`
	// streamtype=pull时，拉流地址
	URL string `json:"url"  gorm:"column:url"`
	// Longitude 经度
	Longitude float64 `json:"longitude" gorm:"column:longitude;type:decimal(10,6)"`
	// Latitude 纬度
	Latitude float64 `json:"latitude" gorm:"column:latitude;type:decimal(10,6)"`
	// GpsTime GPS时间
	GpsTime time.Time `json:"gps_time" gorm:"column:gps_time;type:datetime"`

	addr *sip.Address `gorm:"-"`
}

// 同步摄像头编码格式
func SyncDevicesCodec(ssrc, deviceid string) {
	resp := zlmGetMediaList(zlmGetMediaListReq{streamID: ssrc})
	if resp.Code != 0 {
		logrus.Errorln("syncDevicesCodec fail", ssrc, resp)
		return
	}
	if len(resp.Data) == 0 {
		logrus.Errorln("syncDevicesCodec fail", ssrc, "not found data", resp)
		return
	}
	for _, data := range resp.Data {
		if len(data.Tracks) == 0 {
			logrus.Errorln("syncDevicesCodec fail", ssrc, "not found tracks", resp)
		}

		for _, track := range data.Tracks {
			if track.Type == 0 {
				// 视频
				device := Channels{DeviceID: deviceid}
				if err := db.Get(db.DBClient, &device); err == nil {
					device.VF = transZLMDeviceVF(track.CodecID)
					device.Height = track.Height
					device.Width = track.Width
					device.FPS = int(track.FPS)
					db.Save(db.DBClient, &device)
				} else {
					logrus.Errorln("syncDevicesCodec deviceid not found,deviceid:", deviceid)
				}
			}
		}
	}
}

// 从请求中解析出设备信息
func parserDevicesFromRequest(req *sip.Request) (Devices, bool) {
	u := Devices{}
	header, ok := req.From()
	if !ok {
		logrus.Warningln("not found from header from request", req.String())
		return u, false
	}
	if header.Address == nil {
		logrus.Warningln("not found from user from request", req.String())
		return u, false
	}
	if header.Address.User() == nil {
		logrus.Warningln("not found from user from request", req.String())
		return u, false
	}
	u.DeviceID = header.Address.User().String()
	u.Region = header.Address.Host()
	via, ok := req.ViaHop()
	if !ok {
		logrus.Info("not found ViaHop from request", req.String())
		return u, false
	}
	u.Host = via.Host
	u.Port = via.Port.String()
	report, ok := via.Params.Get("rport")
	if ok && report != nil {
		u.Rport = report.String()
	}
	raddr, ok := via.Params.Get("received")
	if ok && raddr != nil {
		u.RAddr = raddr.String()
	}

	u.TransPort = via.Transport
	u.URIStr = header.Address.String()
	u.addr = sip.NewAddressFromFromHeader(header)
	u.Source = req.Source().String()
	u.source = req.Source()
	return u, true
}

// 获取设备信息（注册设备）
func sipDeviceInfo(to Devices) {
	hb := sip.NewHeaderBuilder().SetTo(to.addr).SetFrom(_serverDevices.addr).AddVia(&sip.ViaHop{
		Params: sip.NewParams().Add("branch", sip.String{Str: sip.GenerateBranch()}),
	}).SetContentType(&sip.ContentTypeXML).SetMethod(sip.MESSAGE)
	req := sip.NewRequest("", sip.MESSAGE, to.addr.URI, sip.DefaultSipVersion, hb.Build(), sip.GetDeviceInfoXML(to.DeviceID))
	req.SetDestination(to.source)
	// 根据设备的传输方式发送请求
	var tx *sip.Transaction
	var err error
	if strings.ToLower(to.TransPort) == "tcp" {
		tx, err = srv.RequestWithProtocol(req, "tcp")
	} else {
		tx, err = srv.Request(req) // 默认UDP
	}
	if err != nil {
		logrus.Warnln("sipDeviceInfo  error,", err)
		return
	}
	_, err = sipResponse(tx)
	if err != nil {
		logrus.Warnln("sipDeviceInfo  response error,", err)
		return
	}
}

// GetActiveDevice 获取活跃设备信息（包含完整的连接信息）
func GetActiveDevice(deviceID string) (Devices, bool) {
	return _activeDevices.Get(deviceID)
}

// UpdateActiveDevice 更新活跃设备的设备信息（用于同步数据库变更到缓存）
func UpdateActiveDevice(deviceID string, device *Devices) {
	if old, ok := _activeDevices.Get(deviceID); ok {
		device.addr = old.addr
		device.source = old.source
		_activeDevices.Store(deviceID, *device)
		logActiveDeviceStore("update_active", device)
	}
}

// SipCatalog 获取注册设备包含的列表 (私有函数，内部实现)
func SipCatalog(to Devices) {
	hb := sip.NewHeaderBuilder().SetTo(to.addr).SetFrom(_serverDevices.addr).AddVia(&sip.ViaHop{
		Params: sip.NewParams().Add("branch", sip.String{Str: sip.GenerateBranch()}),
	}).SetContentType(&sip.ContentTypeXML).SetMethod(sip.MESSAGE)
	req := sip.NewRequest("", sip.MESSAGE, to.addr.URI, sip.DefaultSipVersion, hb.Build(), sip.GetCatalogXML(to.DeviceID))
	req.SetDestination(to.source)

	// 根据设备的连接类型发送请求
	var tx *sip.Transaction
	var err error
	// 根据设备的传输方式发送请求
	if strings.ToLower(to.TransPort) == "tcp" {
		tx, err = srv.RequestWithProtocol(req, "tcp")
	} else {
		tx, err = srv.Request(req) // 默认UDP
	}

	if err != nil {
		logrus.Warnln("SipCatalog  error,", err)
		return
	}
	_, err = sipResponse(tx)
	if err != nil {
		logrus.Warnln("SipCatalog  response error,", err)
		return
	}
}

// sipPTZControl 向设备发送云台控制指令
func SipPTZControl(device Devices, ptzCmd string) error {
	hb := sip.NewHeaderBuilder().
		SetTo(device.addr).
		SetFrom(_serverDevices.addr).
		AddVia(&sip.ViaHop{
			Params: sip.NewParams().Add("branch", sip.String{Str: sip.GenerateBranch()}),
		}).
		SetContentType(&sip.ContentTypeXML).
		SetMethod(sip.MESSAGE)

	// 组装目标地址
	toAddr := device.addr

	req := sip.NewRequest(
		"", sip.MESSAGE, toAddr.URI, sip.DefaultSipVersion, hb.Build(),
		sip.GetPTZControlXML(device.DeviceID, ptzCmd),
	)
	req.SetDestination(device.source)

	var tx *sip.Transaction
	var err error
	if strings.ToLower(device.TransPort) == "tcp" {
		tx, err = srv.RequestWithProtocol(req, "tcp")
	} else {
		tx, err = srv.Request(req)
	}
	if err != nil {
		logrus.Warnln("PTZControl send error,", err)
		return err
	}
	_, err = sipResponse(tx)
	if err != nil {
		logrus.Warnln("PTZControl response error,", err)
		return err
	}
	return nil
}

// MessageDeviceInfoResponse 主设备明细返回结构
type MessageDeviceInfoResponse struct {
	CmdType      string `xml:"CmdType"`
	SN           int    `xml:"SN"`
	DeviceID     string `xml:"DeviceID"`
	DeviceType   string `xml:"DeviceType"`
	Manufacturer string `xml:"Manufacturer"`
	Model        string `xml:"Model"`
	Firmware     string `xml:"Firmware"`
}

func sipMessageDeviceInfo(u Devices, body []byte) error {
	message := &MessageDeviceInfoResponse{}
	if err := utils.XMLDecode([]byte(body), message); err != nil {
		logrus.Errorln("sipMessageDeviceInfo Unmarshal xml err:", err, "body:", body)
		return err
	}
	logrus.Infoln("sipMessageDeviceInfo message:", message)
	existing := Devices{DeviceID: u.DeviceID}
	if err := db.Get(db.DBClient, &existing); err == nil {
		var updates Devices
		updates.ActiveAt = time.Now().Unix()
		if existing.Manufacturer == "" && message.Manufacturer != "" {
			updates.Manufacturer = message.Manufacturer
		}
		if existing.Model == "" && message.Model != "" {
			updates.Model = message.Model
		}
		if existing.DeviceType == "" && message.DeviceType != "" {
			updates.DeviceType = message.DeviceType
		}
		if existing.Firmware == "" && message.Firmware != "" {
			updates.Firmware = message.Firmware
		}
		db.UpdateAll(db.DBClient, new(Devices), db.M{"deviceid=?": u.DeviceID}, updates)
	}
	return nil
}

// MessageDeviceListResponse 设备明细列表返回结构
type MessageDeviceListResponse struct {
	XMLName  xml.Name   `xml:"Response"`
	CmdType  string     `xml:"CmdType"`
	SN       int        `xml:"SN"`
	DeviceID string     `xml:"DeviceID"`
	SumNum   int        `xml:"SumNum"`
	Item     []Channels `xml:"DeviceList>Item"`
}

// sipMessageCatalog 解析Sip中的Catalog信息入库
func sipMessageCatalog(_ Devices, body []byte) error {
	message := &MessageDeviceListResponse{}
	if err := utils.XMLDecode(body, message); err != nil {
		logrus.Errorln("Message Unmarshal xml err:", err, "body:", string(body))
		return err
	}

	if message.SumNum <= 0 {
		return nil
	}

	// 更新同步时间（无论 SN 是多少，只要收到数据就更新）
	catalogSyncState[message.DeviceID] = time.Now().Unix()

	if message.SumNum > 0 {
		for _, d := range message.Item {
			var channel Channels
			var err error
			// 使用 Unscoped 查询，忽略软删除标记，以便找回并更新被删除的通道
			if err = db.DBClient.Unscoped().Where("channelid = ? AND deviceid = ?", d.ChannelID, message.DeviceID).First(&channel).Error; err == nil {
				channel.Active = time.Now().Unix()
				channel.URIStr = fmt.Sprintf("sip:%s@%s", d.ChannelID, _sysinfo.Region)
				channel.Status = transDeviceStatus(d.Status)
				channel.Name = d.Name
				channel.Manufacturer = d.Manufacturer
				channel.Model = d.Model
				channel.Owner = d.Owner
				channel.CivilCode = d.CivilCode
				channel.Address = d.Address
				channel.Parental = d.Parental
				channel.SafetyWay = d.SafetyWay
				channel.RegisterWay = d.RegisterWay
				channel.Secrecy = d.Secrecy
				channel.DeletedAt = nil // 清除软删除标记（复活通道）
				db.Save(db.DBClient, &channel)
				go notify(notifyChannelsActive(channel))
			} else if db.RecordNotFound(err) {
				// 通道不存在，创建新通道
				channel = Channels{
					ChannelID:    d.ChannelID,
					DeviceID:     message.DeviceID,
					Active:       time.Now().Unix(),
					URIStr:       fmt.Sprintf("sip:%s@%s", d.ChannelID, _sysinfo.Region),
					Status:       transDeviceStatus(d.Status),
					Name:         d.Name,
					Manufacturer: d.Manufacturer,
					Model:        d.Model,
					Owner:        d.Owner,
					CivilCode:    d.CivilCode,
					Address:      d.Address,
					Parental:     d.Parental,
					SafetyWay:    d.SafetyWay,
					RegisterWay:  d.RegisterWay,
					Secrecy:      d.Secrecy,
					StreamType:   m.StreamTypePush,
				}
				if err = db.Create(db.DBClient, &channel); err != nil {
					logrus.Errorln("创建通道失败:", err, "channelid:", d.ChannelID, "deviceid:", message.DeviceID)
				} else {
					logrus.Infoln("创建新通道成功:", d.ChannelID, "deviceid:", message.DeviceID)
					go notify(notifyChannelsActive(channel))
				}
			} else {
				logrus.Infoln("deviceid not found,deviceid:", d.DeviceID, "pdid:", message.DeviceID, "err", err)
			}
		}
	}

	// 停止之前的清理定时器
	if timer, exists := catalogSyncTimers[message.DeviceID]; exists {
		timer.Stop()
	}

	// 设置延迟清理任务：12 秒后如果没有新响应，则认为同步完成，执行清理
	catalogSyncTimers[message.DeviceID] = time.AfterFunc(12*time.Second, func() {
		if startTime, ok := catalogSyncState[message.DeviceID]; ok {
			deletedCount := cleanOldChannels(message.DeviceID, startTime)
			logrus.Infoln("[Catalog] 同步完成:", message.DeviceID, "清理旧通道数量:", deletedCount)
			delete(catalogSyncState, message.DeviceID)
			delete(catalogSyncTimers, message.DeviceID)
		}
	})

	return nil
}

// catalogSyncState 记录 Catalog 同步状态
var catalogSyncState = make(map[string]int64)        // deviceID -> syncStartTime
var catalogSyncTimers = make(map[string]*time.Timer) // deviceID -> cleanup timer

// cleanOldChannels 清理指定时间之前更新的通道（设置软删除标记）
func cleanOldChannels(deviceID string, beforeTime int64) int {
	now := time.Now().Unix()
	result := db.DBClient.Model(&Channels{}).
		Where("deviceid = ? AND uptime < ? AND deltime IS NULL", deviceID, beforeTime).
		Update("deltime", now)

	if result.Error != nil {
		logrus.Errorln("[Catalog] 清理旧通道失败:", result.Error)
		return 0
	}

	return int(result.RowsAffected)
}

var deviceStatusMap = map[string]string{
	"ON":     m.DeviceStatusON,
	"OK":     m.DeviceStatusON,
	"ONLINE": m.DeviceStatusON,
	"OFFILE": m.DeviceStatusOFF,
	"OFF":    m.DeviceStatusOFF,
}

func transDeviceStatus(status string) string {
	if v, ok := deviceStatusMap[status]; ok {
		return v
	}
	return status
}

// MobilePosition 设备位置信息
type MobilePosition struct {
	DeviceID  string  `xml:"DeviceID"` // 上报者 ID (通常是 NVR/网关)
	TargetID  string  `xml:"TargetID"` // 目标 ID (具体是哪个设备/通道)
	Time      string  `xml:"Time"`
	Longitude float64 `xml:"Longitude"`
	Latitude  float64 `xml:"Latitude"`
	Speed     float64 `xml:"Speed"`
	Direction float64 `xml:"Direction"`
	Altitude  float64 `xml:"Altitude"`
}

// SipSubscribe 发送订阅请求
func SipSubscribe(device Devices, subscribeType string) {
	if !device.Regist {
		logrus.Debugln("设备未注册，跳过订阅:", device.DeviceID)
		return
	}

	activeDevice, ok := _activeDevices.Get(device.DeviceID)
	if !ok {
		logrus.Debugln("设备不在线，跳过订阅:", device.DeviceID)
		return
	}

	// 构建订阅请求 XML
	xmlBody := fmt.Sprintf(`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>MobilePosition</CmdType>
<SN>%d</SN>
<DeviceID>%s</DeviceID>
<Interval>5</Interval>
</Query>`, utils.RandInt(100000, 999999), device.DeviceID)

	uri, _ := sip.ParseURI(activeDevice.URIStr)
	activeDevice.addr = &sip.Address{URI: uri, Params: sip.NewParams()}

	// 根据设备传输协议设置 Via 的 Transport
	transport := "UDP"
	if strings.ToLower(activeDevice.TransPort) == "tcp" {
		transport = "TCP"
	}

	_serverDevices.addr.Params.Add("tag", sip.String{Str: utils.RandString(20)})
	hb := sip.NewHeaderBuilder().SetToWithParam(activeDevice.addr).SetFrom(_serverDevices.addr).AddVia(&sip.ViaHop{
		Transport: transport,
		Params:    sip.NewParams().Add("branch", sip.String{Str: sip.GenerateBranch()}),
	}).SetContentType(&sip.ContentTypeXML).SetMethod(sip.SUBSCRIBE).SetContact(_serverDevices.addr)

	req := sip.NewRequest("", sip.SUBSCRIBE, activeDevice.addr.URI, sip.DefaultSipVersion, hb.Build(), []byte(xmlBody))
	if activeDevice.source == nil {
		logrus.Warningln("设备连接信息缺失，跳过订阅:", device.DeviceID)
		return
	}
	req.SetDestination(activeDevice.source)
	req.SetRecipient(activeDevice.addr.URI)
	req.AppendHeader(&sip.GenericHeader{HeaderName: "Event", Contents: "presence"})
	req.AppendHeader(&sip.GenericHeader{HeaderName: "Expires", Contents: "3600"})

	var tx *sip.Transaction
	var err error
	if strings.ToLower(activeDevice.TransPort) == "tcp" {
		tx, err = srv.RequestWithProtocol(req, "tcp")
	} else {
		tx, err = srv.Request(req)
	}

	if err != nil {
		logrus.Warningln("发送订阅请求失败:", device.DeviceID, subscribeType, err)
		return
	}

	response, err := sipResponse(tx)
	if err != nil {
		logrus.Warningln("订阅请求响应失败:", device.DeviceID, subscribeType, err)
		return
	}

	logrus.Infoln("订阅成功:", device.DeviceID, subscribeType, response.StatusCode())
}

// CheckAndSubscribe 检查并发送订阅
func CheckAndSubscribe(device Devices) {
	if device.Subscribe == nil {
		return
	}

	// 检查位置订阅
	if pos, ok := device.Subscribe["position"]; ok && pos == true {
		logrus.Infoln("发送位置订阅:", device.DeviceID)
		go SipSubscribe(device, "position")
	}
}
