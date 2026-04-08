package m

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/panjjo/gosip/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config Config
type Config struct {
	MOD       string            `json:"mod" yaml:"mod" mapstructure:"mod"`
	DB        db.Config         `json:"database" yaml:"database" mapstructure:"database"`
	Redis     RedisConfig       `json:"redis" yaml:"redis" mapstructure:"redis"`
	LogLevel  string            `json:"logger" yaml:"logger" mapstructure:"logger"`
	UDP       string            `json:"udp" yaml:"udp" mapstructure:"udp"`
	TCP       string            `json:"tcp" yaml:"tcp" mapstructure:"tcp"`
	API       string            `json:"api" yaml:"api" mapstructure:"api"`
	Secret    string            `json:"secret" yaml:"secret" mapstructure:"secret"`
	Media     MediaServer       `json:"media" yaml:"media" mapstructure:"media"`
	Stream    Stream            `json:"stream" yaml:"stream" mapstructure:"stream"`
	Record    RecordCfg         `json:"record" yaml:"record" mapstructure:"record"`
	GB28181   *SysInfo          `json:"gb28181" yaml:"gb28181" mapstructure:"gb28181"`
	Notify    map[string]string `json:"notify" yaml:"notify" mapstructure:"notify"`
	NotifyMap map[string]string
	JWT       JWTConfig `json:"jwt" yaml:"jwt" mapstructure:"jwt"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `json:"addr" yaml:"addr" mapstructure:"addr"`
	Password string `json:"password" yaml:"password" mapstructure:"password"`
	DB       int    `json:"db" yaml:"db" mapstructure:"db"`
}

type RecordCfg struct {
	FilePath  string `json:"filepath" yaml:"filepath" mapstructure:"filepath"`
	Expire    int    `json:"expire" yaml:"expire"  mapstructure:"expire"`
	Recordmax int    `json:"recordmax" yaml:"recordmax"  mapstructure:"recordmax"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	// TokenExpire Token 有效期（天），默认 7 天
	TokenExpire int `json:"token_expire" yaml:"token_expire" mapstructure:"token_expire"`
}

// Stream Stream
type Stream struct {
	HLS  bool `json:"hls" yaml:"hls" mapstructure:"hls"`
	RTMP bool `json:"rtmp" yaml:"rtmp" mapstructure:"rtmp"`
}

// MediaServer MediaServer
type MediaServer struct {
	RESTFUL string `json:"restful" yaml:"restful" mapstructure:"restful"`
	HTTP    string `json:"http" yaml:"http" mapstructure:"http"`
	WS      string `json:"ws" yaml:"ws" mapstructure:"ws"`
	RTMP    string `json:"rtmp" yaml:"rtmp" mapstructure:"rtmp"`
	RTSP    string `json:"rtsp" yaml:"rtsp" mapstructure:"rtsp"`
	RTP     string `json:"rtp" yaml:"rtp" mapstructure:"rtp"`
	Secret  string `json:"secret" yaml:"secret" mapstructure:"secret"`
}

type SysInfo struct {
	db.DBModel
	// Region 当前域
	Region string `json:"region"   yaml:"region" mapstructure:"region"`
	// CID 通道id固定头部
	CID string `json:"cid"   yaml:"cid" mapstructure:"cid"`
	// CNUM 当前通道数
	CNUM int `json:"cnum" bson:"unum" yaml:"unum" mapstructure:"unum"`
	// DID 设备id固定头部
	DID string `json:"did" bson:"did" yaml:"did" mapstructure:"did"`
	// DNUM 当前设备数
	DNUM int `json:"dnum" bson:"dnum" yaml:"dnum" mapstructure:"dnum"`
	// LID 当前服务id
	LID string `json:"lid" bson:"lid" yaml:"lid" mapstructure:"lid"`
	// PWD 默认设备接入密码
	PWD         string `json:"pwd" yaml:"pwd" mapstructure:"pwd"`
	MediaServer bool
	// 媒体服务器接流地址
	MediaServerRtpIP net.IP `gorm:"-" json:"-"`
	// 媒体服务器接流端口
	MediaServerRtpPort int `gorm:"-"  json:"-"`
}

func DefaultInfo() *SysInfo {
	return MConfig.GB28181
}

var MConfig *Config

func LoadConfig() {
	// 首先设置日志格式化器
	SetupLogger()

	viper.SetConfigType("yml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.SetDefault("logger", "debug")
	viper.SetDefault("udp", "0.0.0.0:5060")
	viper.SetDefault("api", "0.0.0.0:8090")
	viper.SetDefault("mod", "release")
	viper.SetDefault("jwt.token_expire", 7) // 默认 7 天

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatalln("init config error:", err)
	}
	Gb28181Logger.Infoln("init config ok")
	MConfig = &Config{}
	err = viper.Unmarshal(&MConfig)
	if err != nil {
		logrus.Fatalln("init config unmarshal error:", err)
	}
	// logrus.Infof("config :%+v", MConfig)
	level, _ := logrus.ParseLevel(MConfig.LogLevel)
	logrus.SetLevel(level)

	// 设置 Gb28181Logger 的日志级别（从配置文件读取）
	SetLogLevel(MConfig.LogLevel)

	db.DBClient, err = db.Open(MConfig.DB)
	if err != nil {
		logrus.Fatalln("init db error:", err)
	}
	db.DBClient.SetNowFuncOverride(func() interface{} {
		return time.Now().Unix()
	})
	// 配置 GORM 日志：将输出重定向到自定义的 SQL 日志写入器
	sqlWriter := GetSqlLogWriter()
	sqlLog := log.New(sqlWriter, "", 0)
	db.DBClient.SetLogger(sqlLog)
	db.DBClient.LogMode(MConfig.LogLevel == "debug" || MConfig.LogLevel == "trace")
	go db.KeepLive(db.DBClient, time.Minute)

	MConfig.MOD = strings.ToUpper(MConfig.MOD)
	notifyMap := map[string]string{}
	if MConfig.Notify != nil {
		for k, v := range MConfig.Notify {
			if v != "" {
				notifyMap[strings.ReplaceAll(k, "_", ".")] = v
			}
		}
	}
	MConfig.NotifyMap = notifyMap
	if MConfig.Record.Expire <= 0 {
		MConfig.Record.Expire = 7
	}

	if MConfig.Record.Recordmax <= 0 {
		MConfig.Record.Recordmax = 600
	}
}
