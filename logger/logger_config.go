package cherryLogger

import (
	"fmt"
	cfacade "github.com/cherry-game/cherry/facade"
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
		IncludeStdout   bool   `json:"include_stdout"`    // 是否包含os.stdout输出
		IncludeStderr   bool   `json:"include_stderr"`    // 是否包含os.stderr输出
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
		IncludeStdout:   false,
		IncludeStderr:   false,
	}
	return config
}

func NewConfig(jsonConfig cfacade.ProfileJSON) *Config {
	config := &Config{}

	config.Level = jsonConfig.GetString("level", "debug")
	config.StackLevel = jsonConfig.GetString("stack_level", "error")
	config.EnableConsole = jsonConfig.GetBool("enable_console", true)
	config.EnableWriteFile = jsonConfig.GetBool("enable_write_file", false)
	config.MaxAge = jsonConfig.GetInt("max_age", 7)
	config.TimeFormat = jsonConfig.GetString("time_format", "15:04:05.000")
	config.PrintCaller = jsonConfig.GetBool("print_caller", true)
	config.RotationTime = jsonConfig.GetInt("rotation_time", 86400)

	defaultFileLinkPath := fmt.Sprintf("logs/%s.log", config.Level)
	config.FileLinkPath = jsonConfig.GetString("file_link_path", defaultFileLinkPath)
	defaultFilePath := fmt.Sprintf("logs/%s_%s", config.Level, "%Y%m%d%H%M.log")
	config.FilePathFormat = jsonConfig.GetString("file_path_format", defaultFilePath)
	config.IncludeStdout = jsonConfig.GetBool("include_stdout", false)
	config.IncludeStderr = jsonConfig.GetBool("include_stderr", false)

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
		return zapcore.DebugLevel
	}
}
