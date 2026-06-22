package cherryLogger

import (
	"os"
	"strings"
	"time"

	"github.com/cherry-game/cherry/logger/rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// buildEncoder creates a zapcore.Encoder of the requested type and bakes the
// given common fields into it so they appear on every log line automatically.
func buildEncoder(encoderType string, cfg zapcore.EncoderConfig, commonFields map[string]string) zapcore.Encoder {
	var enc zapcore.Encoder
	if strings.EqualFold(encoderType, ENCODER_JSON_Type) {
		enc = zapcore.NewJSONEncoder(cfg)
	} else {
		enc = newKVConsoleEncoder(cfg)
	}
	for k, v := range commonFields {
		enc.AddString(k, v)
	}
	return enc
}

// NewConfigLogger builds a CherryLogger from the given Config. Pass nil for
// commonFields and wrappers when neither is needed; the Manager handles them
// automatically via GetOrCreateLogger.
//
// This function references no package-level global state — it is safe to call
// during package initialization (e.g. from NewManager).
func NewConfigLogger(config Config, commonFields map[string]string, wrappers []Wrapper, opts ...zap.Option) *CherryLogger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		CallerKey:      "caller",
		MessageKey:     "msg",
		NameKey:        "name",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoderConfig.EncodeLevel = func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(level.CapitalString())
	}

	encoderConfig.EncodeTime = config.TimeEncoder()

	if config.PrintCaller {
		encoderConfig.EncodeName = zapcore.FullNameEncoder
		encoderConfig.FunctionKey = zapcore.OmitKey
		opts = append(opts, zap.AddCaller())
	}

	opts = append(opts, zap.AddStacktrace(GetLevel(config.StackLevel)))

	var writers []zapcore.WriteSyncer

	if config.EnableWriteFile {
		hook, err := rotatelogs.New(
			config.FilePathFormat,
			rotatelogs.WithLinkName(config.FileLinkPath),
			rotatelogs.WithMaxAge(time.Hour*24*time.Duration(config.MaxAge)),
			rotatelogs.WithRotationTime(time.Second*time.Duration(config.RotationTime)),
		)
		if err != nil {
			panic(err)
		}
		writers = append(writers, zapcore.AddSync(hook))
	}

	if config.EnableConsole {
		writers = append(writers, zapcore.AddSync(os.Stderr))
	}

	if config.IncludeStdout {
		writers = append(writers, zapcore.Lock(os.Stdout))
	}

	// IncludeStderr is independent of EnableConsole: EnableConsole is always
	// synchronous AddSync, IncludeStderr is locked/buffered. When both are on,
	// include only the locked version to avoid duplicate stderr output.
	if config.IncludeStderr && !config.EnableConsole {
		writers = append(writers, zapcore.Lock(os.Stderr))
	}

	core := zapcore.NewCore(
		buildEncoder(config.EncoderType, encoderConfig, commonFields),
		zapcore.AddSync(zapcore.NewMultiWriteSyncer(writers...)),
		zap.NewAtomicLevelAt(GetLevel(config.LogLevel)),
	)

	for _, v := range wrappers {
		core = v.Wrap(core)
	}

	return &CherryLogger{
		Config:        config,
		SugaredLogger: zap.New(core, opts...).Sugar(),
	}
}
