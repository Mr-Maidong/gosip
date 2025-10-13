package sipapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/panjjo/gosip/db"
	sip "github.com/panjjo/gosip/sip/s"
	"github.com/panjjo/gosip/utils"
	"github.com/sirupsen/logrus"
)

// MessageReceive 接收到的请求数据最外层，主要用来判断数据类型
type MessageReceive struct {
	CmdType string `xml:"CmdType"`
	SN      int    `xml:"SN"`
}

func handlerMessage(req *sip.Request, tx *sip.Transaction) {
	u, ok := parserDevicesFromReqeust(req)
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
	}
	tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
}

func handlerRegister(req *sip.Request, tx *sip.Transaction) {
	// 判断是否存在授权字段
	logrus.Debugln("req:", req)
	if hdrs := req.GetHeaders("Authorization"); len(hdrs) > 0 {
		fromUser, ok := parserDevicesFromReqeust(req)
		if !ok {
			return
		}
		user := Devices{DeviceID: fromUser.DeviceID}
		if err := db.Get(db.DBClient, &user); err == nil {
			if !user.Regist {
				// 如果数据库里用户未激活，替换user数据
				fromUser.ID = user.ID
				fromUser.Name = user.Name
				fromUser.PWD = user.PWD
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
				_activeDevices.Store(user.DeviceID, user)
				if !user.Regist {
					// 第一次激活，保存数据库
					user.Regist = true
					db.DBClient.Save(&user)
					logrus.Infoln("new user regist,id:", user.DeviceID)
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
		if fromUser, ok := parserDevicesFromReqeust(req); ok {
			user := Devices{DeviceID: fromUser.DeviceID}
			if err := db.Get(db.DBClient, &user); err != nil {
				// 设备不存在，发送通知
				logrus.Warnf("未知设备首次注册尝试: DeviceID=%s, Addr=%s", fromUser.DeviceID, fromUser.addr.URI.String())
				go notify(notifyDeviceUnknown(fromUser.DeviceID, fromUser.addr.URI.String()))
				return
			}
		}
	}
	resp := sip.NewResponseFromRequest("", req, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), nil)
	resp.AppendHeader(&sip.GenericHeader{HeaderName: "WWW-Authenticate", Contents: fmt.Sprintf("Digest nonce=\"%s\", algorithm=MD5, realm=\"%s\",qop=\"auth\"", utils.RandString(32), _sysinfo.Region)})
	tx.Respond(resp)
}

// handlerNotify 处理NOTIFY请求
func handlerNotify(req *sip.Request, tx *sip.Transaction) {
	logrus.Debugln("收到NOTIFY请求:", req)

	// 解析设备信息
	fromUser, ok := parserDevicesFromReqeust(req)
	if !ok {
		tx.Respond(sip.NewResponseFromRequest("", req, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil))
		return
	}

	// 检查设备是否已注册
	device, exists := _activeDevices.Get(fromUser.DeviceID)
	if !exists {
		// 设备未注册，返回401
		logrus.Warnf("未注册设备发送NOTIFY: DeviceID=%s", fromUser.DeviceID)
		resp := sip.NewResponseFromRequest("", req, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), nil)
		tx.Respond(resp)
		return
	}

	// 处理NOTIFY消息体（如果有的话）
	if len, have := req.ContentLength(); have && !len.Equals(0) {
		body := req.Body()
		logrus.Debugf("NOTIFY消息体: DeviceID=%s, Body=%s", fromUser.DeviceID, string(body))

		// 这里可以根据需要解析NOTIFY的具体内容
		// 例如：设备状态变化、报警信息等

		// 可以选择性地处理不同类型的NOTIFY消息
		// 目前先简单记录日志
	}

	// 更新设备活跃状态
	device.ActiveAt = time.Now().Unix()
	_activeDevices.Store(fromUser.DeviceID, device)

	// 返回200 OK响应
	tx.Respond(sip.NewResponseFromRequest("", req, http.StatusOK, "OK", nil))

	logrus.Debugf("NOTIFY处理完成: DeviceID=%s", fromUser.DeviceID)
}

// handlerBye 处理BYE请求
func handlerBye(req *sip.Request, tx *sip.Transaction) {
	logrus.Debugln("收到BYE请求:", req)

	// 解析设备信息
	fromUser, ok := parserDevicesFromReqeust(req)
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
	logrus.Debugln("收到OPTIONS请求:", req)

	// 解析设备信息
	fromUser, ok := parserDevicesFromReqeust(req)
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
