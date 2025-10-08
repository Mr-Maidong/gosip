package utils

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// LogSIPMessage 格式化打印SIP消息
func LogSIPMessage(level logrus.Level, prefix string, message string) {
	// 使用自定义格式打印SIP消息
	logrus.StandardLogger().Log(level, fmt.Sprintf("%s message: \n%s", prefix, message))
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
