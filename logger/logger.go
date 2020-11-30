package cherryLogger

import (
	"github.com/cherry-game/cherry/utils"
	json "github.com/json-iterator/go"
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	goPath "path"
	"strings"
	"time"
)

var (
	logger *zap.SugaredLogger

	path         = "log/"
	suffix       = ".log"
	level        = "debug"
	isWriteFile  = false
	maxAge       = 7 //day
	rotationHour = 1
	timeFormat   = "2006-01-02 15:04:05.000"
)

func Logger() *zap.SugaredLogger {
	return logger
}

func DefaultSet() {
	logger = NewConsoleLogger(zapcore.DebugLevel, zap.AddCallerSkip(1))
}

func SetLogger(cfg json.Any) {
	logConfig := cfg.Get("logger")

	if logConfig != nil {
		if logConfig.Get("path") != nil {
			path = logConfig.Get("path").ToString()
		}

		if logConfig.Get("suffix") != nil {
			suffix = logConfig.Get("suffix").ToString()
		}

		if logConfig.Get("level") != nil {
			level = logConfig.Get("level").ToString()
		}

		if logConfig.Get("is_write_file") != nil {
			isWriteFile = logConfig.Get("is_write_file").ToBool()
		}

		if logConfig.Get("max_age") != nil {
			maxAge = logConfig.Get("max_age").ToInt()
		}

		if logConfig.Get("rotation_hour") != nil {
			rotationHour = logConfig.Get("rotation_hour").ToInt()
		}

		if logConfig.Get("time_format") != nil {
			timeFormat = logConfig.Get("time_format").ToString()
		}
	}

	if isWriteFile {
		logger = NewFileLogger(path, level+suffix, maxAge, rotationHour, LogLevel(level), zap.AddCallerSkip(1))
	} else {
		logger = NewConsoleLogger(LogLevel(level), zap.AddCallerSkip(1))
	}
}

func NewConsoleLogger(level zapcore.Level, opts ...zap.Option) *zap.SugaredLogger {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.EncoderConfig.EncodeTime = EncodeTime
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	builder, err := config.Build(opts...)
	if err != nil {
		panic(err)
	}
	return builder.Sugar()
}

func NewFileLogger(filePath, fileName string, maxAge, rotationHour int, level zapcore.Level, opts ...zap.Option) *zap.SugaredLogger {
	cherryUtils.File.CheckPath(filePath)

	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:      "ts",
		LevelKey:     "level",
		MessageKey:   "msg",
		CallerKey:    "file",
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeTime:   EncodeTime,
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})

	return zap.New(
		zapcore.NewTee(
			zapcore.NewCore(encoder,
				zapcore.AddSync(GetWriter(filePath, fileName, maxAge, rotationHour)),
				zap.NewAtomicLevelAt(level),
			),
		), opts...,
	).Sugar()
}

func EncodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(timeFormat))
}

func GetWriter(filePath, fileName string, maxAge, rotationHour int) io.Writer {
	f := goPath.Join(filePath, fileName)
	hook, err := rotateLogs.New(
		f+".%Y%m%d%H%M",
		rotateLogs.WithLinkName(f),
		rotateLogs.WithMaxAge(time.Hour*24*time.Duration(maxAge)),
		rotateLogs.WithRotationTime(time.Minute*time.Duration(rotationHour)), //Hour
	)

	if err != nil {
		panic(err)
	}
	return hook
}

func LogLevel(level string) zapcore.Level {
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
