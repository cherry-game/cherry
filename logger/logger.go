package cherryLogger

import (
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/profile"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

var (
	logger = NewConsoleLogger(NewConsoleConfig())
)

func DefaultLogger() *zap.SugaredLogger {
	return logger
}

func SetNodeLogger(node cherryInterfaces.INode) {
	refLogger := node.Settings().Get("ref_logger").ToString()

	if refLogger == "" {
		logger.Infof("refLogger config not found, used default console logger.")
	}

	//global logger
	logger = NewLogger(refLogger)
}

func NewLogger(refLoggerName string) *zap.SugaredLogger {
	if refLoggerName == "" {
		return nil
	}

	jsonConfig := cherryProfile.Config("logger", refLoggerName)
	config := NewConfig(jsonConfig)

	return NewConsoleLogger(config)
}

func NewConsoleLogger(config *Config, opts ...zap.Option) *zap.SugaredLogger {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		NameKey:        "name",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	if config.PrintTime {
		encoderConfig.TimeKey = "ts"
		encoderConfig.EncodeTime = config.TimeEncoder()
	}

	if config.PrintLevel {
		encoderConfig.LevelKey = "level"
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if config.PrintCaller {
		encoderConfig.CallerKey = "caller"
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		encoderConfig.EncodeName = zapcore.FullNameEncoder
		encoderConfig.FunctionKey = zapcore.OmitKey

		opts = append(opts, zap.AddCaller())
	}

	var writers []zapcore.WriteSyncer
	if config.EnableWriteFile && config.FilePath != "" {
		lumberjack := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSize,
			MaxAge:     config.MaxAge,
			MaxBackups: config.MaxBackups,
			Compress:   config.Compress,
		})
		writers = append(writers, lumberjack)
	}

	if config.EnableConsole {
		writers = append(writers, zapcore.AddSync(os.Stderr))
	}

	opts = append(opts, zap.AddStacktrace(GetLevel(config.StackLevel)))
	opts = append(opts, zap.AddCallerSkip(1))

	level := GetLevel(config.Level)

	return NewSugaredLogger(encoderConfig, zapcore.NewMultiWriteSyncer(writers...), level, opts...)
}

func NewSugaredLogger(
	config zapcore.EncoderConfig,
	writer zapcore.WriteSyncer,
	level zapcore.Level,
	opts ...zap.Option,
) *zap.SugaredLogger {

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(config),
		zapcore.AddSync(writer),
		zap.NewAtomicLevelAt(level),
	)

	zapLogger := zap.New(core, opts...)

	return zapLogger.Sugar()
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

// Warn uses fmt.Sprint to construct and log a message.
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Error uses fmt.Sprint to construct and log a message.
func Error(args ...interface{}) {
	logger.Error(args...)
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func DPanic(args ...interface{}) {
	logger.DPanic(args...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func Panic(args ...interface{}) {
	logger.Panic(args...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
func DPanicf(template string, args ...interface{}) {
	logger.DPanicf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func Panicf(template string, args ...interface{}) {
	logger.Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func Fatalf(template string, args ...interface{}) {
	logger.Fatalf(template, args...)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//  s.With(keysAndValues).Debug(msg)
func Debugw(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Infow(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Warnw(msg string, keysAndValues ...interface{}) {
	logger.Warnw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Errorw(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func DPanicw(msg string, keysAndValues ...interface{}) {
	logger.DPanicw(msg, keysAndValues...)
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func Panicw(msg string, keysAndValues ...interface{}) {
	logger.Panicw(msg, keysAndValues...)
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func Fatalw(msg string, keysAndValues ...interface{}) {
	logger.Fatalw(msg, keysAndValues...)
}
