package utils

import (
	"github.com/sirupsen/logrus"
)

// SIPLoggerHook SIP日志钩子函数，由外部注入
var SIPLoggerHook func(level logrus.Level, prefix string, message string)

// LogSIPMessage 格式化打印SIP消息
func LogSIPMessage(level logrus.Level, prefix string, message string) {
	if SIPLoggerHook != nil {
		SIPLoggerHook(level, prefix, message)
	}
}

// LogSIPRequest 格式化打印SIP请求
func LogSIPRequest(source, method, txKey, message string) {
	LogSIPMessage(logrus.DebugLevel,
		"receive request from: "+source+", method: "+method+", txKey: "+txKey,
		message)
}

// LogSIPResponse 格式化打印SIP响应
func LogSIPResponse(source, txKey, message string) {
	LogSIPMessage(logrus.DebugLevel,
		"receive response from: "+source+", txKey: "+txKey,
		message)
}

// LogSIPSend 格式化打印发送的SIP消息
func LogSIPSend(msgType, destination, txKey, message string) {
	LogSIPMessage(logrus.DebugLevel,
		"send "+msgType+" to: "+destination+", txkey: "+txKey,
		message)
}
