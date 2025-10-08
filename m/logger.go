package m

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// CustomFormatter 自定义日志格式化器
type CustomFormatter struct {
	// TimestampFormat 时间格式
	TimestampFormat string
	// DisableColors 是否禁用颜色
	DisableColors bool
}

// Format 格式化日志条目
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	// 设置默认时间格式
	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = "2006-01-02 15:04:05"
	}

	// 获取调用者信息
	caller := ""
	if entry.HasCaller() {
		funcName := filepath.Base(entry.Caller.Function)
		fileName := filepath.Base(entry.Caller.File)
		caller = fmt.Sprintf(" [%s:%s:%d]", fileName, funcName, entry.Caller.Line)
	}

	// 颜色配置
	var levelColor, resetColor string
	if !f.DisableColors {
		levelColor = getLevelColor(entry.Level)
		resetColor = "\033[0m"
	}

	// 格式化时间戳
	timestamp := entry.Time.Format(timestampFormat)

	// 格式化日志级别
	level := strings.ToUpper(entry.Level.String())

	// 构建基础日志行
	fmt.Fprintf(b, "%s[%s]%s %s%s%s %s",
		levelColor,
		timestamp,
		resetColor,
		levelColor,
		level,
		resetColor,
		caller,
	)

	// 处理消息内容
	message := entry.Message

	// 检查是否是SIP消息（包含 "message: \n" 的日志）
	if strings.Contains(message, "message: \n") {
		parts := strings.Split(message, "message: \n")
		if len(parts) == 2 {
			// 前半部分是描述信息
			fmt.Fprintf(b, " %s\n", parts[0])

			// 后半部分是SIP消息，需要格式化
			sipMessage := strings.TrimSpace(parts[1])
			if sipMessage != "" {
				fmt.Fprintf(b, "%s┌─ SIP Message ─────────────────────────────────────────────────────────────────┐%s\n", levelColor, resetColor)

				// 按行分割SIP消息
				lines := strings.Split(sipMessage, "\n")
				for _, line := range lines {
					line = strings.TrimRight(line, "\r")
					if line == "" {
						fmt.Fprintf(b, "%s│%s\n", levelColor, resetColor)
					} else {
						fmt.Fprintf(b, "%s│%s %s\n", levelColor, resetColor, line)
					}
				}

				fmt.Fprintf(b, "%s└───────────────────────────────────────────────────────────────────────────────┘%s\n", levelColor, resetColor)
			}
		} else {
			fmt.Fprintf(b, " %s\n", message)
		}
	} else {
		// 普通消息
		fmt.Fprintf(b, " %s\n", message)
	}

	return b.Bytes(), nil
}

// getLevelColor 获取日志级别对应的颜色
func getLevelColor(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel, logrus.TraceLevel:
		return "\033[36m" // 青色
	case logrus.InfoLevel:
		return "\033[32m" // 绿色
	case logrus.WarnLevel:
		return "\033[33m" // 黄色
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return "\033[31m" // 红色
	default:
		return "\033[37m" // 白色
	}
}

// SetupLogger 设置日志格式化器
func SetupLogger() {
	// 设置自定义格式化器
	logrus.SetFormatter(&CustomFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   false,
	})

	// 设置报告调用者信息
	logrus.SetReportCaller(true)
}
