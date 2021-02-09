package cherryLogger

import (
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap/zapcore"
	"strings"
	"time"
)

type (
	Config struct {
		Level           string // 输出日志等级
		StackLevel      string // 堆栈输出日志等级
		EnableWriteFile bool   // 是否写文件
		EnableConsole   bool   // 是否输出控制台
		FilePath        string // 日志文件路径
		MaxSize         int    // 单个日志文件最大容量(mb)
		MaxAge          int    // 最大保留天数
		MaxBackups      int    // 最大备份时间
		Compress        bool   // 是否备份压缩
		TimeFormat      string // 时间输出格式
		PrintTime       bool   // 是否打印时间
		PrintLevel      bool   // 是否打印日志等级
		PrintCaller     bool   // 是否打印调用函数
	}
)

func NewConsoleConfig() *Config {
	config := &Config{
		Level:           "debug",
		StackLevel:      "error",
		EnableWriteFile: false,
		EnableConsole:   true,
		FilePath:        "",
		MaxSize:         0,
		MaxAge:          0,
		MaxBackups:      0,
		Compress:        false,
		TimeFormat:      "15:04:05.000", //2006-01-02 15:04:05.000
		PrintTime:       true,
		PrintLevel:      true,
		PrintCaller:     true,
	}
	return config
}

func NewConfig(jsonConfig jsoniter.Any) *Config {
	config := NewConsoleConfig()

	if jsonConfig.LastError() == nil {
		if jsonConfig.Get("level").LastError() == nil {
			config.Level = jsonConfig.Get("level").ToString()
		}

		if jsonConfig.Get("stack_level").LastError() == nil {
			config.StackLevel = jsonConfig.Get("stack_level").ToString()
		}

		if jsonConfig.Get("enable_write_file").LastError() == nil {
			config.EnableWriteFile = jsonConfig.Get("enable_write_file").ToBool()
		}

		if jsonConfig.Get("enable_console").LastError() == nil {
			config.EnableConsole = jsonConfig.Get("enable_console").ToBool()
		}

		if jsonConfig.Get("file_path").LastError() == nil {
			config.FilePath = jsonConfig.Get("file_path").ToString()
		}

		if jsonConfig.Get("max_size").LastError() == nil {
			config.MaxSize = jsonConfig.Get("max_size").ToInt()
		}

		if jsonConfig.Get("max_age").LastError() == nil {
			config.MaxAge = jsonConfig.Get("max_age").ToInt()
		}

		if jsonConfig.Get("max_backups").LastError() == nil {
			config.MaxBackups = jsonConfig.Get("max_backups").ToInt()
		}

		if jsonConfig.Get("compress").LastError() == nil {
			config.Compress = jsonConfig.Get("compress").ToBool()
		}

		if jsonConfig.Get("time_format").LastError() == nil {
			config.TimeFormat = jsonConfig.Get("time_format").ToString()
		}

		if jsonConfig.Get("print_time").LastError() == nil {
			config.PrintTime = jsonConfig.Get("print_time").ToBool()
		}

		if jsonConfig.Get("print_level").LastError() == nil {
			config.PrintLevel = jsonConfig.Get("print_level").ToBool()
		}

		if jsonConfig.Get("print_caller").LastError() == nil {
			config.PrintCaller = jsonConfig.Get("print_caller").ToBool()
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
