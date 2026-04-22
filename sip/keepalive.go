package sipapi

import (
	"context"
	"strings"
	"time"

	"github.com/panjjo/gosip/db"
	"github.com/panjjo/gosip/utils"
	"github.com/sirupsen/logrus"
)

const (
	DeviceKeyPrefix = "device:"
)

type MessageNotify struct {
	CmdType  string `xml:"CmdType"`
	SN       int    `xml:"SN"`
	DeviceID string `xml:"DeviceID"`
	Status   string `xml:"Status"`
	Info     string `xml:"Info"`
}

func sipMessageKeepalive(u Devices, body []byte) error {
	message := &MessageNotify{}
	if err := utils.XMLDecode(body, message); err != nil {
		logrus.Errorln("Message Unmarshal xml err:", err, "body:", string(body))
		return err
	}
	device, ok := _activeDevices.Get(u.DeviceID)
	logrus.Debugln("Device Keepalive:", u.DeviceID, message.Status)
	if !ok {
		device = Devices{DeviceID: u.DeviceID}
		if err := db.Get(db.DBClient, &device); err != nil {
			logrus.Warnln("Device Keepalive not found ", u.DeviceID, err)
		}
	}
	if message.Status == "OK" {
		// 只更新 SIP 消息中实际存在的字段，其他字段保留原值
		// 需要从 SIP 消息更新的字段
		if u.Host != "" {
			device.Host = u.Host
		}
		if u.Port != "" {
			device.Port = u.Port
		}
		if u.Rport != "" {
			device.Rport = u.Rport
		}
		if u.RAddr != "" {
			device.RAddr = u.RAddr
		}
		if u.TransPort != "" {
			device.TransPort = u.TransPort
		}
		if u.URIStr != "" {
			device.URIStr = u.URIStr
		}
		if u.addr != nil {
			device.addr = u.addr
		}
		if u.Source != "" {
			device.Source = u.Source
		}
		if u.source != nil {
			device.source = u.source
		}
		// 保留以下字段（这些字段不在 SIP 消息中）
		// Regist、Online、StreamIP、SipIP、Name、PWD 等保持原值
		device.Online = true
		device.ActiveAt = time.Now().Unix()
		_activeDevices.Store(u.DeviceID, device)
		logActiveDeviceStore("keepalive_ok", &device)
		if db.RedisClient != nil {
			if err := db.RefreshDeviceRedis(u.DeviceID); err != nil {
				logrus.Warnln("Refresh device redis error:", u.DeviceID, err)
			}
		}
	} else {
		device.ActiveAt = -1
		_activeDevices.Delete(u.DeviceID)
		logActiveDeviceDelete("keepalive_fail", u.DeviceID)
		if db.RedisClient != nil {
			if err := db.DeleteDeviceRedis(u.DeviceID); err != nil {
				logrus.Warnln("Delete device redis error:", u.DeviceID, err)
			}
		}
	}
	go notify(notifyDevicesActive(u.DeviceID, message.Status))
	_, err := db.UpdateAll(db.DBClient, new(Devices), map[string]any{"deviceid=?": u.DeviceID}, Devices{
		Host:     u.Host,
		Port:     u.Port,
		Rport:    u.Rport,
		RAddr:    u.RAddr,
		Source:   u.Source,
		URIStr:   u.URIStr,
		Online:   true,
		ActiveAt: device.ActiveAt,
	})
	return err
}

func StartDeviceOfflineWatcher(ctx context.Context) {
	if db.RedisClient == nil {
		logrus.Warnln("Redis client not initialized, skip device offline watcher")
		return
	}
	go func() {
		logrus.Infoln("Starting device offline watcher...")
		pubsub := db.RedisClient.PSubscribe(ctx, "__keyevent@0__:expired")
		defer pubsub.Close()
		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				logrus.Infoln("Device offline watcher stopped")
				return
			case msg := <-ch:
				if msg == nil {
					continue
				}
				key := msg.Payload
				if after, ok := strings.CutPrefix(key, DeviceKeyPrefix); ok {
					deviceID := after
					logrus.Infof("Device offline detected: %s", deviceID)
					err := db.DBClient.Model(&Devices{}).Where("deviceid = ?", deviceID).Update("online", false).Error
					if err != nil {
						logrus.Errorln("Update device offline error:", deviceID, err)
					} else {
						db.DBClient.Create(&db.DeviceEvent{
							DeviceID:  deviceID,
							EventType: "OFFLINE",
							EventTime: time.Now().Unix(),
						})
						_activeDevices.Delete(deviceID)
						logActiveDeviceDelete("offline_watcher", deviceID)
						go notify(notifyDevicesActive(deviceID, "OFFLINE"))
					}
				}
			}
		}
	}()
}
