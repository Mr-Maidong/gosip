package m

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var Gb28181Logger *logrus.Logger
var sqlLogFile *os.File

// sqlLogWriter 自定义的 io.Writer，用于格式化 GORM 日志
type sqlLogWriter struct {
	file *os.File
}

func (w *sqlLogWriter) Write(p []byte) (n int, err error) {
	message := string(p)

	// 检查是否是 GORM 的 SQL 日志消息
	if strings.Contains(message, "C:/") || strings.Contains(message, "D:/") {
		// 解析并格式化 GORM 日志
		formatted := formatGormLog(message)
		return w.file.Write([]byte(formatted))
	}

	// 默认直接写入
	return w.file.Write(p)
}

// formatGormLog 格式化 GORM 日志
func formatGormLog(message string) string {
	var b bytes.Buffer

	// 提取调用位置（多行显示）
	var locations []string
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 寻找包含文件路径和行号的行
		if strings.Contains(line, ".go:") {
			locations = append(locations, line)
		}
	}
	callLocation := ""
	if len(locations) > 0 {
		callLocation = strings.Join(locations, "\n│   ")
	}

	// 提取 SQL 和耗时
	// GORM v1 日志格式多变，可能是:
	// [time] [1.23ms] SQL ...
	// 1.23ms SQL ...
	// 1.23msSQL ... (无空格)

	duration := ""
	sql := ""

	// 正则匹配：可选括号 + 数字 + 单位 + 可选空格 + SQL 语句
	// SQL 通常以 SELECT, UPDATE, INSERT, DELETE 等开头
	sqlRegex := regexp.MustCompile(`\[?(\d+(?:\.\d+)?(?:ms|µs|ns))\]?\s*([A-Z].*)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过堆栈路径
		if strings.HasPrefix(line, "C:/") || strings.HasPrefix(line, "D:/") || strings.HasPrefix(line, "(") || strings.HasPrefix(line, ")") {
			continue
		}

		match := sqlRegex.FindStringSubmatch(line)
		if len(match) > 2 {
			duration = match[1]
			rawSql := match[2]

			// 提取 SQL，去掉后面的 [values] rows
			// values 通常以 [ 开头
			if idx := strings.Index(rawSql, "["); idx != -1 {
				sql = strings.TrimSpace(rawSql[:idx])
			} else {
				sql = rawSql
			}

			// 清理末尾的 rows affected 数字 (如果是单独的数字)
			sql = strings.TrimSuffix(sql, "0")
			sql = strings.TrimSuffix(sql, "1")
			sql = strings.TrimSpace(sql)

			if sql != "" {
				break // 找到有效的 SQL 就停止
			}
		}
	}

	// 格式化输出
	fmt.Fprintf(&b, "┌─ SQL ─────────────────────────────────────────────────────────────────┐\n")
	if callLocation != "" {
		fmt.Fprintf(&b, "│ Location: \n│   %s\n", callLocation)
	}
	if duration != "" {
		fmt.Fprintf(&b, "│ Duration: %s\n", duration)
	}

	if sql != "" {
		fmt.Fprintf(&b, "│ SQL: %s\n", sql)
	} else {
		// 如果没解析出 SQL，打印原始内容的一小部分以便调试
		fmt.Fprintf(&b, "│ (未解析到 SQL，原始日志片段:)\n")
		for _, line := range lines {
			l := strings.TrimSpace(line)
			if l != "" && !strings.HasPrefix(l, "C:/") && !strings.HasPrefix(l, "D:/") && l != "(" && l != ")" {
				fmt.Fprintf(&b, "│   %s\n", l)
			}
		}
	}
	fmt.Fprintf(&b, "───────────────────────────────────────────────────────────────────────\n\n")

	return b.String()
}

func init() {
	// 注意：此时配置文件尚未加载，设置一个默认级别（Trace）确保所有日志都能被记录
	// LoadConfig() 执行后会根据配置文件覆盖此设置
	
	// 使用绝对路径，适配测试环境
	logDir := "logs"
	if absPath, err := filepath.Abs(logDir); err == nil {
		logDir = absPath
	}
	os.MkdirAll(logDir, 0755)
	
	logFile, err := os.OpenFile(filepath.Join(logDir, "gb28181.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("无法打开日志文件: " + err.Error())
	}
	Gb28181Logger = logrus.New()
	Gb28181Logger.Out = logFile
	Gb28181Logger.SetLevel(logrus.TraceLevel)
	Gb28181Logger.SetFormatter(&CustomFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   true,
	})

	// 初始化 SQL 日志文件
	sqlLogFile, err = os.OpenFile(filepath.Join(logDir, "sql.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("无法打开 SQL 日志文件: " + err.Error())
	}
}

// GetSqlLogWriter 获取 SQL 日志写入器
func GetSqlLogWriter() *sqlLogWriter {
	return &sqlLogWriter{file: sqlLogFile}
}

// SetLogLevel 根据配置文件设置日志级别
func SetLogLevel(level string) {
	if Gb28181Logger == nil {
		return
	}
	switch strings.ToLower(level) {
	case "trace":
		Gb28181Logger.SetLevel(logrus.TraceLevel)
	case "debug":
		Gb28181Logger.SetLevel(logrus.DebugLevel)
	case "info":
		Gb28181Logger.SetLevel(logrus.InfoLevel)
	case "warn":
		Gb28181Logger.SetLevel(logrus.WarnLevel)
	case "error":
		Gb28181Logger.SetLevel(logrus.ErrorLevel)
	default:
		Gb28181Logger.SetLevel(logrus.DebugLevel)
	}
}

// LogSIPMessage 格式化打印SIP消息
func LogSIPMessage(level logrus.Level, prefix string, message string) {
	Gb28181Logger.Log(level, fmt.Sprintf("%s message: \n%s", prefix, message))
}

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
				fmt.Fprintf(b, "%s─ SIP Message ─────────────────────────────────────────────────────────────────%s\n", levelColor, resetColor)

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

				fmt.Fprintf(b, "%s────────────────────────────────────────────────────────────────────────────%s\n", levelColor, resetColor)
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
