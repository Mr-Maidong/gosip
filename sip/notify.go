package sipapi

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/panjjo/gosip/utils"
	"github.com/sirupsen/logrus"
)

const (
	// NotifyMethodDeviceUnknown 未知设备尝试注册通知
	NotifyMethodDeviceUnknown = "devices.unknown"
	// NotifyMethodUserActive 设备活跃状态通知
	NotifyMethodDevicesActive = "devices.active"
	// NotifyMethodUserRegister 设备注册通知
	NotifyMethodDevicesRegister = "devices.regiester"
	// NotifyMethodDeviceActive 通道活跃通知
	NotifyMethodChannelsActive = "channels.active"
	// NotifyMethodRecordStop 视频录制结束
	NotifyMethodRecordStop = "records.stop"
)

// Notify 消息通知结构
type Notify struct {
	Method string      `json:"method"`
	Data   interface{} `json:"data"`
}

func notify(data *Notify) {
	if url, ok := config.NotifyMap[data.Method]; ok {
		res, err := utils.PostJSONRequest(url, data)
		if err != nil {
			logrus.Warningln(data.Method, "send notify fail.", err)
		}
		if strings.ToUpper(string(res)) != "OK" {
			logrus.Warningln(data.Method, "send notify resp fail.", string(res), "len:", len(res), config.Notify, data)
		} else {
			logrus.Debug("notify send succ:", data.Method, data.Data)
		}
	} else {
		logrus.Traceln("notify config not found", data.Method)
	}
}

func notifyDeviceUnknown(deviceID, addr string) *Notify {
	return &Notify{
		Method: NotifyMethodDeviceUnknown,
		Data: map[string]any{
			"deviceid": deviceID,
			"addr":     addr,
			"time":     time.Now().Unix(),
			"message":  "未知设备尝试注册，请在控制台添加该设备",
		},
	}
}

func notifyDevicesAcitve(id, status string) *Notify {
	return &Notify{
		Method: NotifyMethodDevicesActive,
		Data: map[string]any{
			"deviceid": id,
			"status":   status,
			"time":     time.Now().Unix(),
		},
	}
}
func notifyDevicesRegister(u Devices) *Notify {
	u.Sys = *config.GB28181
	return &Notify{
		Method: NotifyMethodDevicesRegister,
		Data:   u,
	}
}

func notifyChannelsActive(d Channels) *Notify {
	return &Notify{
		Method: NotifyMethodChannelsActive,
		Data: map[string]any{
			"channelid": d.ChannelID,
			"status":    d.Status,
			"time":      time.Now().Unix(),
		},
	}
}
func notifyRecordStop(url string, req url.Values) *Notify {
	d := map[string]any{
		"url": fmt.Sprintf("%s/%s", config.Media.HTTP, url),
	}
	for k, v := range req {
		d[k] = v[0]
	}
	return &Notify{
		Method: NotifyMethodRecordStop,
		Data:   d,
	}
}
