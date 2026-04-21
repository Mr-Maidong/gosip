package sipapi

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/panjjo/gosip/db"
	"github.com/panjjo/gosip/m"
	sip "github.com/panjjo/gosip/sip/s"
	"github.com/panjjo/gosip/utils"
	"github.com/sirupsen/logrus"
)

func Start() {
	// 数据库表初始化 启动时自动同步数据结构到数据库
	db.DBClient.AutoMigrate(new(Devices))
	db.DBClient.AutoMigrate(new(Channels))
	db.DBClient.AutoMigrate(new(Streams))
	db.DBClient.AutoMigrate(new(m.SysInfo))
	db.DBClient.AutoMigrate(new(Files))

	LoadSYSInfo()

	srv = sip.NewServer()
	srv.RegistHandler(sip.OPTIONS, handlerOptions)
	srv.RegistHandler(sip.MESSAGE, handlerMessage)
	srv.RegistHandler(sip.REGISTER, handlerRegister)
	srv.RegistHandler(sip.NOTIFY, handlerNotify)
	srv.RegistHandler(sip.BYE, handlerBye)
	go srv.ListenTCPServer(config.TCP)
	go srv.ListenUDPServer(config.UDP)

	// 注册TCP连接断开回调
	utils.TCPConnCloseHook = handleTCPConnClose
}

// handleTCPConnClose TCP连接断开时处理
// 根据断开连接的远程地址查找设备并标记为离线
func handleTCPConnClose(remoteAddr string) {
	_activeDevices.Range(func(key, value interface{}) bool {
		device := value.(Devices)
		if device.Source == remoteAddr {
			logrus.Infof("TCP连接断开，标记设备离线: DeviceID=%s, Addr=%s", device.DeviceID, remoteAddr)
			// 先更新数据库
			if err := db.DBClient.Model(&Devices{}).Where("deviceid = ?", device.DeviceID).Update("online", false).Error; err != nil {
				logrus.Errorln("TCP断开更新设备状态失败:", device.DeviceID, err)
			}
			// 删除Redis
			if db.RedisClient != nil {
				if err := db.DeleteDeviceRedis(device.DeviceID); err != nil {
					logrus.Errorln("TCP断开删除Redis失败:", device.DeviceID, err)
				}
			}
			// 发送离线通知
			go notify(notifyDevicesActive(device.DeviceID, "OFFLINE"))
			// 最后从缓存中删除
			_activeDevices.Delete(device.DeviceID)
			logActiveDeviceDelete("tcp_close", device.DeviceID)
			return false // 匹配到设备后停止遍历
		}
		return true
	})
}

// ActiveDevices 记录当前活跃设备，请求播放时设备必须处于活跃状态
type ActiveDevices struct {
	sync.Map
}

// Get Get
func (a *ActiveDevices) Get(key string) (Devices, bool) {
	if v, ok := a.Load(key); ok {
		return v.(Devices), ok
	}
	return Devices{}, false
}

var _activeDevices ActiveDevices

// 系统运行信息
var _sysinfo *m.SysInfo
var config *m.Config

func LoadSYSInfo() {

	config = m.MConfig
	_activeDevices = ActiveDevices{sync.Map{}}

	StreamList = streamsList{&sync.Map{}, &sync.Map{}, 0}
	ssrcLock = &sync.Mutex{}
	_recordList = &sync.Map{}
	RecordList = apiRecordList{items: map[string]*apiRecordItem{}, l: sync.RWMutex{}}

	// init sysinfo
	_sysinfo = &m.SysInfo{}
	if err := db.Get(db.DBClient, _sysinfo); err != nil {
		if db.RecordNotFound(err) {
			//  初始不存在
			_sysinfo = m.DefaultInfo()

			if err = db.Create(db.DBClient, _sysinfo); err != nil {
				logrus.Fatalf("1 init sysinfo err:%v", err)
			}
		} else {
			logrus.Fatalf("2 init sysinfo err:%v", err)
		}
	}
	m.MConfig.GB28181 = _sysinfo

	uri, _ := sip.ParseSipURI(fmt.Sprintf("sip:%s@%s", _sysinfo.LID, _sysinfo.Region))
	_serverDevices = Devices{
		DeviceID: _sysinfo.LID,
		Region:   _sysinfo.Region,
		addr: &sip.Address{
			DisplayName: sip.String{Str: "sipserver"},
			URI:         &uri,
			Params:      sip.NewParams(),
		},
	}

	// init media
	url, err := url.Parse(config.Media.RTP)
	if err != nil {
		logrus.Fatalf("media rtp url error,url:%s,err:%v", config.Media.RTP, err)
	}
	ipaddr, err := net.ResolveIPAddr("ip", url.Hostname())
	if err != nil {
		logrus.Fatalf("media rtp url error,url:%s,err:%v", config.Media.RTP, err)
	}
	_sysinfo.MediaServerRtpIP = ipaddr.IP
	_sysinfo.MediaServerRtpPort, _ = strconv.Atoi(url.Port())
}

// 新增函数：基于 deviceId 和 channelId 生成 StreamID
func generateStreamID(deviceID, channelID string) string {
	return fmt.Sprintf("live_%s_%s", deviceID, channelID)
}

// 生成回放流ID
func generateReplayStreamID(deviceID, channelID string, startTime int64) string {
	return fmt.Sprintf("replay_%s_%s_%d", deviceID, channelID, startTime)
}

// 生成对讲流ID
func generateTalkStreamID(deviceID, channelID string) string {
	return fmt.Sprintf("talk_%s_%s", deviceID, channelID)
}

// zlm接收到的ssrc为16进制。发起请求的ssrc为10进制
func ssrc2stream(ssrc string) string {
	if ssrc[0:1] == "0" {
		ssrc = ssrc[1:]
	}
	num, _ := strconv.Atoi(ssrc)
	return fmt.Sprintf("%08X", num)
}

func sipResponse(tx *sip.Transaction) (*sip.Response, error) {
	response := tx.GetResponse()
	if response == nil {
		return nil, utils.NewError(nil, "response timeout", "tx key:", tx.Key())
	}
	if response.StatusCode() != http.StatusOK {
		return response, utils.NewError(nil, "response fail", response.StatusCode(), response.Reason(), "tx key:", tx.Key())
	}
	return response, nil
}
