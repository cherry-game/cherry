package cherryLogger

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap/zapcore"
	"strings"
	"time"
)

type (
	Config struct {
		Level           string `json:"level"`             // 输出日志等级
		StackLevel      string `json:"stack_level"`       // 堆栈输出日志等级
		EnableConsole   bool   `json:"enable_console"`    // 是否控制台输出
		EnableWriteFile bool   `json:"enable_write_file"` // 是否输出文件(必需配置FilePath)
		MaxAge          int    `json:"max_age"`           // 最大保留天数(达到限制，则会被清理)
		TimeFormat      string `json:"time_format"`       // 打印时间输出格式
		PrintCaller     bool   `json:"print_caller"`      // 是否打印调用函数
		RotationTime    int    `json:"rotation_time"`     // 日期分割时间(秒)
		FileLinkPath    string `json:"file_link_path"`    // 日志文件连接路径
		FilePathFormat  string `json:"file_path_format"`  // 日志文件路径格式
	}
)

func defaultConsoleConfig() *Config {
	config := &Config{
		Level:           "debug",
		StackLevel:      "error",
		EnableConsole:   true,
		EnableWriteFile: false,
		MaxAge:          7,
		TimeFormat:      "15:04:05.000", //2006-01-02 15:04:05.000
		PrintCaller:     true,
		RotationTime:    86400,
		FileLinkPath:    "logs/log.log",
		FilePathFormat:  "logs/log_%Y%m%d%H%M.log",
	}
	return config
}

func NewConfig(jsonConfig jsoniter.Any) *Config {
	config := defaultConsoleConfig()

	if jsonConfig.LastError() == nil {
		if jsonConfig.Get("level").LastError() == nil {
			config.Level = jsonConfig.Get("level").ToString()
		}

		if jsonConfig.Get("stack_level").LastError() == nil {
			config.StackLevel = jsonConfig.Get("stack_level").ToString()
		}

		if jsonConfig.Get("enable_console").LastError() == nil {
			config.EnableConsole = jsonConfig.Get("enable_console").ToBool()
		}

		if jsonConfig.Get("enable_write_file").LastError() == nil {
			config.EnableWriteFile = jsonConfig.Get("enable_write_file").ToBool()
		}

		if jsonConfig.Get("max_age").LastError() == nil {
			config.MaxAge = jsonConfig.Get("max_age").ToInt()
		}

		if jsonConfig.Get("time_format").LastError() == nil {
			config.TimeFormat = jsonConfig.Get("time_format").ToString()
		}

		if jsonConfig.Get("print_caller").LastError() == nil {
			config.PrintCaller = jsonConfig.Get("print_caller").ToBool()
		}

		if jsonConfig.Get("rotation_time").LastError() == nil {
			config.RotationTime = jsonConfig.Get("rotation_time").ToInt()
		}

		if jsonConfig.Get("file_link_path").LastError() == nil {
			config.FileLinkPath = jsonConfig.Get("file_link_path").ToString()
		} else {
			config.FileLinkPath = fmt.Sprintf("logs/%s.log", config.Level)
		}

		if jsonConfig.Get("file_path_format").LastError() == nil {
			config.FilePathFormat = jsonConfig.Get("file_path_format").ToString()
		} else {
			config.FilePathFormat = fmt.Sprintf("logs/%s_%s", config.Level, "%Y%m%d%H%M.log")
		}
	}

	return config
}

func (c *Config) TimeEncoder() zapcore.TimeEncoder {
	return func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(time.Format(c.TimeFormat))
	}
}

func GetLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.FatalLevel
	}
}
