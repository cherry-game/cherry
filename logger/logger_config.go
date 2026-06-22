package cherryLogger

import (
	"fmt"
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
	cprofile "github.com/cherry-game/cherry/profile"
	"go.uber.org/zap/zapcore"
)

// Config describes the construction parameters for a CherryLogger. It is
// typically populated from a JSON profile (via NewConfig / NewConfigWithName).
//
// After the logger is built, the Config is stored on CherryLogger as a
// read-only snapshot — mutating it has no effect on the running logger.
type Config struct {
	// LogLevel is the minimum enabled log level ("debug", "info", etc.).
	LogLevel string `json:"level"`
	// StackLevel is the minimum level at which stack traces are captured.
	StackLevel string `json:"stack_level"`
	// EncoderType selects the encoder: "console" (key=value) or "json".
	EncoderType string `json:"encoder_type"`
	// EnableConsole enables synchronous writes to os.Stderr.
	EnableConsole bool `json:"enable_console"`
	// EnableWriteFile enables file output with rotation via rotatelogs.
	EnableWriteFile bool `json:"enable_write_file"`
	// MaxAge is the maximum number of days a rotated log file is kept.
	MaxAge int `json:"max_age"`
	// TimeFormat is the layout passed to time.Format for timestamps.
	TimeFormat string `json:"time_format"`
	// PrintCaller enables the caller file:line annotation on every log line.
	PrintCaller bool `json:"print_caller"`
	// RotationTime is the file rotation interval in seconds.
	RotationTime int `json:"rotation_time"`
	// FileLinkPath is the fixed symlink / shortcut name pointing to the current log file.
	FileLinkPath string `json:"file_link_path"`
	// FilePathFormat is the rotated file path pattern (strftime syntax).
	FilePathFormat string `json:"file_path_format"`
	// IncludeStdout adds os.Stdout as an extra writer.
	IncludeStdout bool `json:"include_stdout"`
	// IncludeStderr adds os.Stderr as an extra writer (skipped when EnableConsole is on).
	IncludeStderr bool `json:"include_stderr"`
}

// defaultConfig provides baseline values used by defaultConsoleConfig and as
// fallback defaults in NewConfig.
var defaultConfig = Config{
	LogLevel:        "debug",                     // emit debug and above
	StackLevel:      "error",                     // capture stack traces at error and above
	EncoderType:     "console",                   // key=value text format
	EnableConsole:   true,                        // write to stderr
	EnableWriteFile: false,                       // no file output by default
	MaxAge:          7,                           // keep rotated files for 7 days
	TimeFormat:      "15:04:05.000",              // time with milliseconds
	PrintCaller:     true,                        // include caller file:line
	RotationTime:    86400,                       // rotate every 86400 seconds (1 day)
	FileLinkPath:    "logs/debug.log",            // symlink to current log file
	FilePathFormat:  "logs/debug_%Y%m%d%H%M.log", // rotated file naming pattern
	IncludeStdout:   false,                       // no extra stdout writer
	IncludeStderr:   false,                       // no extra stderr writer (already covered by EnableConsole)
}

// defaultConsoleConfig returns a copy of the default console-only Config.
func defaultConsoleConfig() Config {
	return defaultConfig
}

// NewConfig creates a Config from a JSON profile node. Every field falls back
// to the value in defaultConfig when the profile key is absent.
func NewConfig(jsonConfig cfacade.ProfileJSON) (*Config, error) {
	config := Config{
		LogLevel:        jsonConfig.GetString("level", defaultConfig.LogLevel),
		StackLevel:      jsonConfig.GetString("stack_level", defaultConfig.StackLevel),
		EncoderType:     jsonConfig.GetString("encoder_type", defaultConfig.EncoderType),
		EnableConsole:   jsonConfig.GetBool("enable_console", defaultConfig.EnableConsole),
		EnableWriteFile: jsonConfig.GetBool("enable_write_file", defaultConfig.EnableWriteFile),
		MaxAge:          jsonConfig.GetInt("max_age", defaultConfig.MaxAge),
		TimeFormat:      jsonConfig.GetString("time_format", defaultConfig.TimeFormat),
		PrintCaller:     jsonConfig.GetBool("print_caller", defaultConfig.PrintCaller),
		RotationTime:    jsonConfig.GetInt("rotation_time", defaultConfig.RotationTime),
		FileLinkPath:    jsonConfig.GetString("file_link_path", defaultConfig.FileLinkPath),
		FilePathFormat:  jsonConfig.GetString("file_path_format", defaultConfig.FilePathFormat),
		IncludeStdout:   jsonConfig.GetBool("include_stdout", defaultConfig.IncludeStdout),
		IncludeStderr:   jsonConfig.GetBool("include_stderr", defaultConfig.IncludeStderr),
	}

	if config.EnableWriteFile {
		if config.FileLinkPath == "" {
			config.FileLinkPath = fmt.Sprintf("logs/%s.log", config.LogLevel)
		}

		if config.FilePathFormat == "" {
			config.FilePathFormat = fmt.Sprintf("logs/%s_%s.log", config.LogLevel, "%Y%m%d%H%M")
		}
	}

	return &config, nil
}

// NewConfigWithName looks up a named logger definition under the "logger" key
// in the global profile and returns its Config.
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

// TimeEncoder returns a zapcore.TimeEncoder that formats timestamps using the
// Config's TimeFormat layout.
func (c *Config) TimeEncoder() zapcore.TimeEncoder {
	return func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(time.Format(c.TimeFormat))
	}
}
