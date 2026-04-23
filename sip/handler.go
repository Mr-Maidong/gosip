package sipapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/panjjo/gosip/db"
	sip "github.com/panjjo/gosip/sip/s"
	"github.com/panjjo/gosip/utils"
	"github.com/sirupsen/logrus"
)

const gbTimeLayout = "2006-01-02T15:04:05"

// MessageReceive 接收到的请求数据最外层，主要用来判断数据类型
type MessageReceive struct {
	CmdType string `xml:"CmdType"`
	SN      int    `xml:"SN"`
}

func handlerMessage(req *sip.Request, tx *sip.Transaction) {
	u, ok := parserDevicesFromRequest(req)
	if !ok {
		// 未解析出来源用户返回错误
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
		return
	}
	// 判断是否存在body数据
	if len, have := req.ContentLength(); !have || len.Equals(0) {
		// 不存在就直接返回的成功
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
		return
	}
	body := req.Body()
	message := &MessageReceive{}

	if err := utils.XMLDecode(body, message); err != nil {
		logrus.Warnln("Message Unmarshal xml err:", err, "body:", string(body))
		// 有些body xml发送过来的不带encoding ，而且格式不是utf8的，导致xml解析失败，此处使用gbk转utf8后再次尝试xml解析
		body, err = utils.GbkToUtf8(body)
		if err != nil {
			logrus.Errorln("message gbk to utf8 err", err)
		}
		if err := utils.XMLDecode(body, message); err != nil {
			logrus.Errorln("Message Unmarshal xml after gbktoutf8 err:", err, "body:", string(body))
			tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
			return
		}
	}
	switch message.CmdType {
	case "Catalog":
		// 设备列表
		sipMessageCatalog(u, body)
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
		return
	case "Keepalive":
		// heardbeat
		if err := sipMessageKeepalive(u, body); err == nil {
			tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
			return
		}
	case "RecordInfo":
		// 设备音视频文件列表
		sipMessageRecordInfo(u, body)
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
	case "DeviceInfo":
		// 主设备信息
		sipMessageDeviceInfo(u, body)
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
		return
	default:
		// 打印未处理的消息类型，用于调试
		logrus.Warnf("[Notify] 收到未处理的消息类型: CmdType=%s, DeviceID=%s, Body=%s",
			message.CmdType, u.DeviceID, string(body))
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
		return
	}
	tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
}

func handlerRegister(req *sip.Request, tx *sip.Transaction) {
	// 判断是否存在授权字段
	if hdrs := req.GetHeaders("Authorization"); len(hdrs) > 0 {
		fromUser, ok := parserDevicesFromRequest(req)
		if !ok {
			return
		}
		user := Devices{DeviceID: fromUser.DeviceID}
		logrus.Infoln("查询数据库用户信息，DeviceID:", fromUser.DeviceID)
		if err := db.Get(db.DBClient, &user); err == nil {
			if !user.Regist {
				fromUser.ID = user.ID
				fromUser.Name = user.Name
				fromUser.PWD = user.PWD
				fromUser.Manufacturer = user.Manufacturer
				fromUser.Model = user.Model
				fromUser.DeviceType = user.DeviceType
				fromUser.Firmware = user.Firmware
				user = fromUser
			}
			user.addr = fromUser.addr
			authenticateHeader := hdrs[0].(*sip.GenericHeader)
			auth := sip.AuthFromValue(authenticateHeader.Contents)
			auth.SetPassword(user.PWD)
			auth.SetUsername(user.DeviceID)
			auth.SetMethod(string(req.Method()))
			auth.SetURI(auth.Get("uri"))
			if auth.CalcResponse() == auth.Get("response") {
				// 验证成功
				// 记录活跃设备
				user.source = fromUser.source
				user.addr = fromUser.addr
				user.Online = true
				_activeDevices.Store(user.DeviceID, user)
				logActiveDeviceStore("handler_register", &user)
				// 记录设备上线事件
				db.DBClient.Create(&db.DeviceEvent{
					DeviceID:  user.DeviceID,
					EventType: "ONLINE",
					EventTime: time.Now().Unix(),
					Source:   user.Source,
				})
				// 确保 Online 状态同步到数据库
				db.DBClient.Model(&Devices{}).Where("deviceid = ?", user.DeviceID).Update("online", true)
				if !user.Regist {
					// 第一次激活，保存数据库
					user.Regist = true
					db.DBClient.Save(&user)
					logrus.Infoln("new user regist,id:", user.DeviceID)
					// 检查并发送订阅
					go CheckAndSubscribe(user)
				}
				if db.RedisClient != nil {
					if err := db.RefreshDeviceRedis(user.DeviceID); err != nil {
						logrus.Warnln("Refresh device redis error:", user.DeviceID, err)
					}
				}
				tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))
				// 注册成功后查询设备信息，获取制作厂商等信息
				go notify(notifyDevicesRegister(user))
				go sipDeviceInfo(fromUser)
				return
			}
		} else {
			// 设备不存在于数据库中，发送通知提醒管理员
			logrus.Warnf("未知设备尝试注册: DeviceID=%s, Addr=%s", fromUser.DeviceID, fromUser.addr.URI.String())
			go notify(notifyDeviceUnknown(fromUser.DeviceID, fromUser.addr.URI.String()))
			return
		}
	} else {
		// 首次注册请求（无Authorization头），解析设备信息并记录
		if fromUser, ok := parserDevicesFromRequest(req); ok {
			user := Devices{DeviceID: fromUser.DeviceID}
			if err := db.Get(db.DBClient, &user); err != nil {
				// 设备不存在，自动插入数据库（使用配置的默认密码）
				logrus.Infof("自动注册新设备: DeviceID=%s, Addr=%s", fromUser.DeviceID, fromUser.addr.URI.String())
				newUser := fromUser
				newUser.Name = fromUser.DeviceID
				newUser.PWD = _sysinfo.PWD
				if err := db.Create(db.DBClient, &newUser); err != nil {
					logrus.Errorf("自动注册设备失败: DeviceID=%s, err=%v", fromUser.DeviceID, err)
					go notify(notifyDeviceUnknown(fromUser.DeviceID, fromUser.addr.URI.String()))
					return
				}
				if db.RedisClient != nil {
					if err := db.RefreshDeviceRedis(newUser.DeviceID); err != nil {
						logrus.Warnln("Refresh device redis error:", newUser.DeviceID, err)
					}
				}
			}
		}
	}
	resp := sip.NewResponseFromRequest("", req, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), nil)
	resp.AppendHeader(&sip.GenericHeader{HeaderName: "WWW-Authenticate", Contents: fmt.Sprintf("Digest nonce=\"%s\", algorithm=MD5, realm=\"%s\",qop=\"auth\"", utils.RandString(32), _sysinfo.Region)})
	tx.Respond(resp)
}

// handlerNotify 处理NOTIFY请求
func handlerNotify(req *sip.Request, tx *sip.Transaction) {
	// 解析设备信息
	fromUser, ok := parserDevicesFromRequest(req)
	if !ok {
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
		return
	}

	// 处理NOTIFY消息体（优先处理位置上报，不依赖注册状态）
	if len, have := req.ContentLength(); have && !len.Equals(0) {
		body := req.Body()
		// 解析 MobilePosition 消息
		parseMobilePosition(fromUser.DeviceID, body)
	}

	// 检查设备是否已注册
	device, exists := _activeDevices.Get(fromUser.DeviceID)
	if !exists {
		// 设备未注册，返回200 OK（位置已处理）
		logrus.Debugf("设备未注册，但已处理位置上报: DeviceID=%s", fromUser.DeviceID)
		resp := sip.NewResponseFromRequest("", req, http.StatusOK, http.StatusText(http.StatusOK), nil)
		tx.Respond(resp)
		return
	}

	// 更新设备活跃状态
	device.ActiveAt = time.Now().Unix()
	_activeDevices.Store(fromUser.DeviceID, device)
	logActiveDeviceStore("handler_notify", &device)

	// 返回200 OK响应
	tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))

	logrus.Debugf("NOTIFY处理完成: DeviceID=%s", fromUser.DeviceID)
}

// parseMobilePosition 解析设备位置信息
func parseMobilePosition(deviceID string, body []byte) {
	pos := &MobilePosition{}
	if err := utils.XMLDecode(body, pos); err != nil {
		// 尝试 GBK 转 UTF-8 后再次解析
		body, err = utils.GbkToUtf8(body)
		if err != nil {
			logrus.Debugln("MobilePosition gbk to utf8 fail:", err)
			return
		}
		if err := utils.XMLDecode(body, pos); err != nil {
			logrus.Debugln("MobilePosition parse fail:", err)
			return
		}
	}

	// 判断归属 ID：优先使用 TargetID，如果没有则使用 DeviceID
	locationID := pos.TargetID
	if locationID == "" {
		locationID = pos.DeviceID
	}

	gpsTime, err := time.ParseInLocation(gbTimeLayout, pos.Time, time.Local)
	if err != nil {
		logrus.Warnln("解析GPS时间失败:", pos.Time, err)
	}

	// 缓存位置到 Redis（实时写入）
	db.SetDevicePosition(locationID, &db.CachedPosition{
		Longitude: pos.Longitude,
		Latitude:  pos.Latitude,
		GPSTime:   pos.Time,
		Speed:    pos.Speed,
		Direction: pos.Direction,
		Altitude: pos.Altitude,
	})

	// 写入位置历史表（device_positions）
	var devicePositionID string

	// 1. 尝试在通道表中查找
	channel := Channels{ChannelID: locationID}
	if err := db.Get(db.DBClient, &channel); err == nil {
		logrus.Infoln("收到【通道】位置上报:", locationID,
			"经度:", pos.Longitude,
			"纬度:", pos.Latitude,
			"时间:", pos.Time)
		if err := db.DBClient.Model(&Channels{}).Where("channelid = ?", locationID).Updates(map[string]interface{}{
			"longitude": pos.Longitude,
			"latitude":  pos.Latitude,
			"gps_time":   gpsTime,
		}).Error; err != nil {
			logrus.Errorln("更新通道位置信息失败:", locationID, err)
		}
		devicePositionID = locationID
	} else {
		// 2. 尝试在设备表中查找
		device := Devices{DeviceID: locationID}
		if err := db.Get(db.DBClient, &device); err == nil {
			logrus.Infoln("收到【设备】位置上报:", locationID,
				"经度:", pos.Longitude,
				"纬度:", pos.Latitude,
				"时间:", pos.Time)
			if err := db.DBClient.Model(&Devices{}).Where("deviceid = ?", locationID).Updates(map[string]interface{}{
				"longitude": pos.Longitude,
				"latitude":  pos.Latitude,
				"gps_time":   gpsTime,
			}).Error; err != nil {
				logrus.Errorln("更新设备位置信息失败:", locationID, err)
			}
			devicePositionID = locationID
		} else {
			logrus.Warnln("收到位置上报，但未找到对应主体:", locationID)
			return
		}
	}

	// 写入位置历史表
	if devicePositionID != "" && gpsTime.Unix() > 0 {
		devicePos := db.DevicePosition{
			DeviceID:  devicePositionID,
			Longitude: pos.Longitude,
			Latitude:  pos.Latitude,
			GPSTime:   gpsTime.Unix(),
			Speed:     pos.Speed,
			Direction: pos.Direction,
			Altitude:  pos.Altitude,
			CreatedAt: time.Now().Unix(),
		}
		if err := db.Create(db.DBClient, &devicePos); err != nil {
			logrus.Errorln("写入位置历史失败:", devicePositionID, err)
		} else {
			logrus.Debugln("写入位置历史成功:", devicePositionID, gpsTime.Unix())
		}
	}
}

// handlerBye 处理BYE请求
func handlerBye(req *sip.Request, tx *sip.Transaction) {
	// 解析设备信息
	fromUser, ok := parserDevicesFromRequest(req)
	if !ok {
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
		return
	}

	// 检查设备是否已注册
	_, exists := _activeDevices.Get(fromUser.DeviceID)
	if !exists {
		// 设备未注册，返回401
		logrus.Warnf("未注册设备发送BYE: DeviceID=%s", fromUser.DeviceID)
		resp := sip.NewResponseFromRequest("", req, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), nil)
		tx.Respond(resp)
		return
	}

	// 处理BYE请求 - 通常用于结束会话
	// 可以根据Call-ID查找并清理相关的流媒体会话
	if callID, ok := req.CallID(); ok {
		logrus.Infof("设备 %s 请求结束会话: CallID=%s", fromUser.DeviceID, string(*callID))

		// 查找并停止相关流
		go func() {
			// 遍历活跃流，找到匹配的CallID并停止
			StreamList.Response.Range(func(key, value interface{}) bool {
				if stream, ok := value.(*Streams); ok {
					if stream.CallID == string(*callID) {
						logrus.Infof("找到匹配的流，准备停止: StreamID=%s, CallID=%s", stream.StreamID, stream.CallID)
						// 调用停止流的函数
						SipStopPlay(stream.StreamID)
					}
				}
				return false // 找到匹配项后停止遍历
			})
		}()
	}

	// 返回200 OK响应
	tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))

	logrus.Debugf("BYE处理完成: DeviceID=%s", fromUser.DeviceID)
}

// handlerOptions 处理OPTIONS请求
func handlerOptions(req *sip.Request, tx *sip.Transaction) {
	// 解析设备信息
	fromUser, ok := parserDevicesFromRequest(req)
	if !ok {
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
		return
	}

	// 返回支持的方法列表
	resp := sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil)
	resp.AppendHeader(&sip.GenericHeader{
		HeaderName: "Allow",
		Contents:   "REGISTER, MESSAGE, NOTIFY, BYE, OPTIONS, INFO, INVITE, ACK, CANCEL",
	})
	tx.Respond(resp)

	logrus.Debugf("OPTIONS处理完成: DeviceID=%s", fromUser.DeviceID)
}

// StartPositionSyncWorker 定时同步GPS缓存到数据库
func StartPositionSyncWorker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	logrus.Infoln("GPS位置同步worker已启动(30s间隔)")

	for {
		select {
		case <-ctx.Done():
			logrus.Infoln("GPS位置同步worker已停止")
			return
		case <-ticker.C:
			syncGPSPositionsToDB()
		}
	}
}

// syncGPSPositionsToDB 同步GPS缓存到数据库
func syncGPSPositionsToDB() {
	keys, err := db.GetAllGPSKeys()
	if err != nil {
		logrus.Warnln("获取GPS缓存key失败:", err)
		return
	}
	if len(keys) == 0 {
		return
	}

	var positions []db.DevicePosition
	now := time.Now().Unix()

	for _, key := range keys {
		deviceID := key[13:]
		pos, err := db.GetDevicePosition(deviceID)
		if err != nil || pos == nil {
			continue
		}
		gpsTime, _ := time.ParseInLocation(gbTimeLayout, pos.GPSTime, time.Local)
		positions = append(positions, db.DevicePosition{
			DeviceID:   deviceID,
			Longitude: pos.Longitude,
			Latitude:  pos.Latitude,
			GPSTime:   gpsTime.Unix(),
			Speed:    pos.Speed,
			Direction: pos.Direction,
			Altitude:  pos.Altitude,
			CreatedAt: now,
		})
	}

	if len(positions) > 0 {
		if err := db.CreateBatch(db.DBClient, positions); err != nil {
			logrus.Warnln("批量写入GPS历史失败:", err)
		} else {
			logrus.Debugf("同步GPS历史记录 %d 条", len(positions))
		}
	}
}
