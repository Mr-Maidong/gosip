package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	gb28181Logger *logrus.Logger
	loggerOnce    sync.Once
)

// getGb28181Logger 获取GB28181专用日志记录器（单例）
func getGb28181Logger() *logrus.Logger {
	loggerOnce.Do(func() {
		gb28181Logger = logrus.New()

		// 确保 logs 目录存在
		logDir := "logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			logrus.Errorln("创建 logs 目录失败:", err)
			return
		}

		logFile, err := os.OpenFile(filepath.Join(logDir, "gb28181.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logrus.Errorln("打开 gb28181.log 失败:", err)
			return
		}

		gb28181Logger.Out = logFile
		gb28181Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			DisableColors:   true,
		})
		gb28181Logger.SetLevel(logrus.TraceLevel) // 记录所有级别的消息
	})

	return gb28181Logger
}

// LogSIPMessage 格式化打印SIP消息
func LogSIPMessage(level logrus.Level, prefix string, message string) {
	logger := getGb28181Logger()
	if logger != nil {
		logger.Log(level, fmt.Sprintf("%s message: \n%s", prefix, message))
	}
}

// LogSIPRequest 格式化打印SIP请求
func LogSIPRequest(source, method, txKey, message string) {
	LogSIPMessage(logrus.TraceLevel,
		fmt.Sprintf("receive request from: %s, method: %s, txKey: %s", source, method, txKey),
		message)
}

// LogSIPResponse 格式化打印SIP响应
func LogSIPResponse(source, txKey, message string) {
	LogSIPMessage(logrus.TraceLevel,
		fmt.Sprintf("receive response from: %s, txKey: %s", source, txKey),
		message)
}

// LogSIPSend 格式化打印发送的SIP消息
func LogSIPSend(msgType, destination, txKey, message string) {
	LogSIPMessage(logrus.TraceLevel,
		fmt.Sprintf("send %s to: %s, txkey: %s", msgType, destination, txKey),
		message)
}
