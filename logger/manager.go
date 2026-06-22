package cherryLogger

import (
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Wrapper allows intercepting and modifying the zapcore.Core during logger
// construction. Useful for adding sampling, filtering, or metrics to every
// logger built by a Manager without modifying the builder itself.
type Wrapper interface {
	Wrap(core zapcore.Core) zapcore.Core
}

// Manager holds all logger state that was previously package-level globals.
// A default Manager (defaultManager) owns the DefaultLogger used by the
// package-level convenience functions.
//
// For isolated contexts (e.g. tests, multiple Application instances), create a
// separate Manager with NewManager and pass it around explicitly.
type Manager struct {
	mu       sync.RWMutex
	loggers  map[string]*CherryLogger
	wrappers []Wrapper

	defaultLogger *CherryLogger

	commonFields map[string]string
	fileNameVars map[string]string
	printLevel   zapcore.Level
}

// ManagerOption is a functional option for NewManager.
type ManagerOption func(m *Manager)

// WithCommonFields sets the initial common fields for a new Manager.
func WithCommonFields(fields map[string]string) ManagerOption {
	return func(m *Manager) {
		for k, v := range fields {
			m.commonFields[k] = v
		}
	}
}

// NewManager creates a Manager with an initial console-only default logger.
func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{
		loggers:      make(map[string]*CherryLogger),
		commonFields: make(map[string]string),
		fileNameVars: make(map[string]string),
		printLevel:   zapcore.DebugLevel,
	}
	for _, opt := range opts {
		opt(m)
	}
	m.defaultLogger = NewConfigLogger(defaultConsoleConfig(), m.commonFields, m.wrappers, zap.AddCallerSkip(1))
	return m
}

// DefaultLogger returns the manager's default logger (console fallback).
func (m *Manager) DefaultLogger() *CherryLogger {
	return m.defaultLogger
}

// Loggers returns a snapshot of all named loggers in the registry.
func (m *Manager) Loggers() map[string]*CherryLogger {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]*CherryLogger, len(m.loggers))
	for k, v := range m.loggers {
		out[k] = v
	}
	return out
}

// GetOrCreateLogger returns an existing named logger or creates one from profile config.
func (m *Manager) GetOrCreateLogger(refLoggerName string, opts ...zap.Option) *CherryLogger {
	if refLoggerName == "" {
		return m.defaultLogger
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if logger, found := m.loggers[refLoggerName]; found {
		return logger
	}

	config, err := NewConfigWithName(refLoggerName)
	if err != nil {
		Panicf("New Config fail. err = %v", err)
	}

	logger := m.buildWithConfig(config, opts...)
	m.loggers[refLoggerName] = logger
	return logger
}

// buildWithConfig injects fileName vars, common fields, and wrappers,
// then delegates to the shared buildLoggerCore. Caller must hold m.mu.
func (m *Manager) buildWithConfig(config *Config, opts ...zap.Option) *CherryLogger {
	if config.EnableWriteFile {
		for key, value := range m.fileNameVars {
			config.FileLinkPath = strings.ReplaceAll(config.FileLinkPath, "%"+key, value)
			config.FilePathFormat = strings.ReplaceAll(config.FilePathFormat, "%"+key, value)
		}
	}

	return NewConfigLogger(*config, m.commonFields, m.wrappers, opts...)
}

// SetCommonField sets a single key-value pair applied to every log line.
func (m *Manager) SetCommonField(key, value string) {
	m.mu.Lock()
	m.commonFields[key] = value
	m.mu.Unlock()
}

// SetCommonFields sets multiple key-value pairs at once.
func (m *Manager) SetCommonFields(fields map[string]string) {
	m.mu.Lock()
	for k, v := range fields {
		m.commonFields[k] = v
	}
	m.mu.Unlock()
}

// CommonFields returns a copy of the current common fields.
func (m *Manager) CommonFields() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]string, len(m.commonFields))
	for k, v := range m.commonFields {
		out[k] = v
	}
	return out
}

// SetFileNameVar sets a template variable for file path substitution.
func (m *Manager) SetFileNameVar(key, value string) {
	m.mu.Lock()
	m.fileNameVars[key] = value
	m.mu.Unlock()
}

// RegisterWrapper adds a core wrapper that is applied to every logger built by this manager.
func (m *Manager) RegisterWrapper(w Wrapper) {
	m.mu.Lock()
	m.wrappers = append(m.wrappers, w)
	m.mu.Unlock()
}

// SetPrintLevel updates the minimum print level.
func (m *Manager) SetPrintLevel(level zapcore.Level) {
	m.mu.Lock()
	m.printLevel = level
	m.mu.Unlock()
}

// PrintLevel returns true if the given level meets the minimum print threshold.
func (m *Manager) PrintLevel(level zapcore.Level) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return level >= m.printLevel
}

// Sync flushes all loggers managed by this Manager.
func (m *Manager) Sync() {
	_ = m.defaultLogger.Sync()

	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, logger := range m.loggers {
		_ = logger.Sync()
	}
}

// RegisterWrapper adds a core wrapper to the default manager.
func RegisterWrapper(w Wrapper) {
	defaultManager.RegisterWrapper(w)
}

// defaultManager is the package-level singleton used by DefaultLogger and the
// global convenience functions (Info, Debug, etc.). It preserves the
// pre-refactor API for callers that don't need isolation.
var defaultManager = NewManager()
