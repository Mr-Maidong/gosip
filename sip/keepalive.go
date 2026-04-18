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
		// 更新时保留 StreamIP、SipIP 和 Regist（这些字段不在心跳消息中）
		oldStreamIP := device.StreamIP
		oldSipIP := device.SipIP
		oldRegist := device.Regist
		device = u
		device.StreamIP = oldStreamIP
		device.SipIP = oldSipIP
		device.Regist = oldRegist
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
	go notify(notifyDevicesAcitve(u.DeviceID, message.Status))
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
						_activeDevices.Delete(deviceID)
						logActiveDeviceDelete("offline_watcher", deviceID)
						go notify(notifyDevicesAcitve(deviceID, "OFFLINE"))
					}
				}
			}
		}
	}()
}
