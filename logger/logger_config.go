package cherryLogger

import (
	"fmt"
	cprofile "github.com/cherry-game/cherry/profile"
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
	"go.uber.org/zap/zapcore"
)

type (
	Config struct {
		LogLevel        string `json:"level"`             // 输出日志等级
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
		LogLevel:        "debug",
		StackLevel:      "error",
		EnableConsole:   true,
		EnableWriteFile: false,
		MaxAge:          7,
		TimeFormat:      "15:04:05.000", //2006-01-02 15:04:05.000
		PrintCaller:     true,
		RotationTime:    86400,
		FileLinkPath:    "logs/debug.log",
		FilePathFormat:  "logs/debug_%Y%m%d%H%M.log",
		IncludeStdout:   false,
		IncludeStderr:   false,
	}
	return config
}

func NewConfig(jsonConfig cfacade.ProfileJSON) (*Config, error) {
	config := &Config{
		LogLevel:        jsonConfig.GetString("level", "debug"),
		StackLevel:      jsonConfig.GetString("stack_level", "error"),
		EnableConsole:   jsonConfig.GetBool("enable_console", true),
		EnableWriteFile: jsonConfig.GetBool("enable_write_file", false),
		MaxAge:          jsonConfig.GetInt("max_age", 7),
		TimeFormat:      jsonConfig.GetString("time_format", "15:04:05.000"),
		PrintCaller:     jsonConfig.GetBool("print_caller", true),
		RotationTime:    jsonConfig.GetInt("rotation_time", 86400),
		FileLinkPath:    jsonConfig.GetString("file_link_path", ""),
		FilePathFormat:  jsonConfig.GetString("file_path_format", ""),
		IncludeStdout:   jsonConfig.GetBool("include_stdout", false),
		IncludeStderr:   jsonConfig.GetBool("include_stderr", false),
	}

	if config.EnableWriteFile {
		if config.FileLinkPath == "" {
			defaultValue := fmt.Sprintf("logs/%s.log", config.LogLevel)
			config.FileLinkPath = jsonConfig.GetString("file_link_path", defaultValue)
		}

		if config.FilePathFormat == "" {
			defaultValue := fmt.Sprintf("logs/%s_%s.log", config.LogLevel, "%Y%m%d%H%M")
			config.FilePathFormat = jsonConfig.GetString("file_path_format", defaultValue)
		}
	}

	return config, nil
}

func NewConfigWithName(refLoggerName string) (*Config, error) {
	loggerConfig := cprofile.GetConfig("logger")
	if loggerConfig.LastError() != nil {
		return nil, loggerConfig.LastError()
	}

	jsonConfig := loggerConfig.GetConfig(refLoggerName)
	if jsonConfig.LastError() != nil {
		return nil, jsonConfig.LastError()
	}

	return NewConfig(jsonConfig)
}

func (c *Config) TimeEncoder() zapcore.TimeEncoder {
	return func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(time.Format(c.TimeFormat))
	}
}
