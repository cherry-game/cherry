package cherryLogger

import (
	"strings"

	cfacade "github.com/cherry-game/cherry/facade"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// KEY_NODE_TYPE is the common-field key for the node type (e.g. "game", "gate").
	KEY_NODE_TYPE = "nodetype"
	// KEY_NODE_ID is the common-field key for the node's unique ID.
	KEY_NODE_ID = "nodeid"
	// ENCODER_JSON_Type selects the JSON encoder when passed to a Config.
	ENCODER_JSON_Type = "json"
)

// DefaultLogger is the package-level default logger backed by defaultManager.
// It is always a valid console-logging CherryLogger, never nil.
var DefaultLogger = defaultManager.DefaultLogger()

// CherryLogger wraps zap.SugaredLogger with framework-level Config awareness.
// It embeds *zap.SugaredLogger (all logging methods) and a value copy of Config
// that reflects the logger's construction-time settings.
type CherryLogger struct {
	Config
	*zap.SugaredLogger
}

// SetNodeLogger reads the node's profile, determines the referenced logger
// name, and replaces DefaultLogger with a fully configured instance (file
// writer, common fields, fileName vars) from the profile config.
func SetNodeLogger(node cfacade.INode) {
	refLoggerName := node.Settings().Get("ref_logger").ToString()
	if refLoggerName == "" {
		DefaultLogger.Warnf("RefLoggerName not found, used default console logger.")
		return
	}

	defaultManager.SetFileNameVar(KEY_NODE_TYPE, node.NodeType())
	defaultManager.SetFileNameVar(KEY_NODE_ID, node.NodeID())

	defaultManager.SetCommonField(KEY_NODE_TYPE, node.NodeType())
	defaultManager.SetCommonField(KEY_NODE_ID, node.NodeID())

	DefaultLogger = defaultManager.GetOrCreateLogger(refLoggerName, zap.AddCallerSkip(1))
}

// SetFileNameVar sets a template variable on the default manager. Keys such as
// "nodetype" or "nodeid" can be used in log file paths via %key placeholders.
func SetFileNameVar(key, value string) {
	defaultManager.SetFileNameVar(key, value)
}

// SetCommonField sets a single common field on the default manager. Common
// fields appear on every log line produced by the manager's loggers.
func SetCommonField(key, value string) {
	defaultManager.SetCommonField(key, value)
}

// SetCommonFields sets multiple common fields at once on the default manager.
func SetCommonFields(fields map[string]string) {
	defaultManager.SetCommonFields(fields)
}

// CommonFields returns a copy of the default manager's current common fields.
func CommonFields() map[string]string {
	return defaultManager.CommonFields()
}

// Flush syncs all loggers in the default manager. Call before shutdown to
// ensure buffered log entries are written.
func Flush() {
	defaultManager.Sync()
}

// NewLogger creates or retrieves a named logger from profile config via the
// default manager. Loggers are cached by name; subsequent calls with the same
// name return the existing instance.
func NewLogger(refLoggerName string, opts ...zap.Option) *CherryLogger {
	return defaultManager.GetOrCreateLogger(refLoggerName, opts...)
}

// Enable reports whether the given log level is enabled on DefaultLogger.
func Enable(level zapcore.Level) bool {
	return DefaultLogger.Desugar().Core().Enabled(level)
}

// PrintLevel returns true if the given level meets the minimum print threshold
// set on the default manager.
func PrintLevel(level zapcore.Level) bool {
	return defaultManager.PrintLevel(level)
}

// SetPrintLevel updates the minimum print threshold on the default manager.
func SetPrintLevel(level zapcore.Level) {
	defaultManager.SetPrintLevel(level)
}

// GetLevel converts a level name string to a zapcore.Level. Supported values
// (case-insensitive): "debug", "info", "warn", "error", "panic", "fatal".
// Unknown values default to DebugLevel.
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

// --- Package-level convenience functions ---
// Each family mirrors zap.SugaredLogger: plain (fmt.Sprint), f (fmt.Sprintf),
// and w (key=value pairs). All delegate to DefaultLogger.

// Debug logs a message at DebugLevel.
func Debug(args ...interface{}) { DefaultLogger.Debug(args...) }

// Info logs a message at InfoLevel.
func Info(args ...interface{}) { DefaultLogger.Info(args...) }

// Warn logs a message at WarnLevel.
func Warn(args ...interface{}) { DefaultLogger.Warn(args...) }

// Error logs a message at ErrorLevel.
func Error(args ...interface{}) { DefaultLogger.Error(args...) }

// DPanic logs a message at DPanicLevel. In development mode the logger then panics.
func DPanic(args ...interface{}) { DefaultLogger.DPanic(args...) }

// Panic logs a message at PanicLevel, then panics.
func Panic(args ...interface{}) { DefaultLogger.Panic(args...) }

// Fatal logs a message at FatalLevel, then calls os.Exit(1).
func Fatal(args ...interface{}) { DefaultLogger.Fatal(args...) }

// Debugf formats and logs a message at DebugLevel.
func Debugf(template string, args ...interface{}) { DefaultLogger.Debugf(template, args...) }

// Infof formats and logs a message at InfoLevel.
func Infof(template string, args ...interface{}) { DefaultLogger.Infof(template, args...) }

// Warnf formats and logs a message at WarnLevel.
func Warnf(template string, args ...interface{}) { DefaultLogger.Warnf(template, args...) }

// Errorf formats and logs a message at ErrorLevel.
func Errorf(template string, args ...interface{}) { DefaultLogger.Errorf(template, args...) }

// DPanicf formats and logs a message at DPanicLevel. In development mode the logger then panics.
func DPanicf(template string, args ...interface{}) { DefaultLogger.DPanicf(template, args...) }

// Panicf formats and logs a message at PanicLevel, then panics.
func Panicf(template string, args ...interface{}) { DefaultLogger.Panicf(template, args...) }

// Fatalf formats and logs a message at FatalLevel, then calls os.Exit(1).
func Fatalf(template string, args ...interface{}) { DefaultLogger.Fatalf(template, args...) }

// Debugw logs a message at DebugLevel with key-value pairs.
func Debugw(msg string, keysAndValues ...interface{}) { DefaultLogger.Debugw(msg, keysAndValues...) }

// Infow logs a message at InfoLevel with key-value pairs.
func Infow(msg string, keysAndValues ...interface{}) { DefaultLogger.Infow(msg, keysAndValues...) }

// Warnw logs a message at WarnLevel with key-value pairs.
func Warnw(msg string, keysAndValues ...interface{}) { DefaultLogger.Warnw(msg, keysAndValues...) }

// Errorw logs a message at ErrorLevel with key-value pairs.
func Errorw(msg string, keysAndValues ...interface{}) { DefaultLogger.Errorw(msg, keysAndValues...) }

// DPanicw logs a message at DPanicLevel with key-value pairs. In development mode the logger then panics.
func DPanicw(msg string, keysAndValues ...interface{}) { DefaultLogger.DPanicw(msg, keysAndValues...) }

// Panicw logs a message at PanicLevel with key-value pairs, then panics.
func Panicw(msg string, keysAndValues ...interface{}) { DefaultLogger.Panicw(msg, keysAndValues...) }

// Fatalw logs a message at FatalLevel with key-value pairs, then calls os.Exit(1).
func Fatalw(msg string, keysAndValues ...interface{}) { DefaultLogger.Fatalw(msg, keysAndValues...) }
