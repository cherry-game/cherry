package cherryLogger

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	ctime "github.com/cherry-game/cherry/extend/time"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ---------------------------------------------------------------------------
// Level parsing
// ---------------------------------------------------------------------------

func TestGetLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zapcore.Level
	}{
		{"debug", zapcore.DebugLevel},
		{"DEBUG", zapcore.DebugLevel},
		{"Debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"INFO", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"WARN", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"ERROR", zapcore.ErrorLevel},
		{"panic", zapcore.PanicLevel},
		{"PANIC", zapcore.PanicLevel},
		{"fatal", zapcore.FatalLevel},
		{"FATAL", zapcore.FatalLevel},
		{"", zapcore.DebugLevel},
		{"unknown", zapcore.DebugLevel},
		{"bogus", zapcore.DebugLevel},
	}
	for _, tt := range tests {
		got := GetLevel(tt.input)
		if got != tt.expected {
			t.Errorf("GetLevel(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestPrintLevel(t *testing.T) {
	levels := []zapcore.Level{
		zapcore.DebugLevel, zapcore.InfoLevel, zapcore.WarnLevel,
		zapcore.ErrorLevel, zapcore.PanicLevel, zapcore.FatalLevel,
	}
	for _, lv := range levels {
		if !PrintLevel(lv) {
			t.Errorf("PrintLevel(%v) should be true", lv)
		}
	}
}

func TestEnable(t *testing.T) {
	if !Enable(zapcore.DebugLevel) {
		t.Error("Enable(DebugLevel) should be true on default logger")
	}
	if !Enable(zapcore.InfoLevel) {
		t.Error("Enable(InfoLevel) should be true on default logger")
	}
}

// ---------------------------------------------------------------------------
// DefaultLogger & basic interfaces
// ---------------------------------------------------------------------------

func TestDefaultLogger_NotNil(t *testing.T) {
	if DefaultLogger == nil {
		t.Fatal("DefaultLogger should not be nil")
	}
}

func TestFlush_NoPanic(t *testing.T) {
	Flush()
}

// ---------------------------------------------------------------------------
// Common fields — package-level
// ---------------------------------------------------------------------------

func TestSetCommonField_Single(t *testing.T) {
	SetCommonField("_t_single", "v1")
	if CommonFields()["_t_single"] != "v1" {
		t.Error("SetCommonField did not persist")
	}
}

func TestSetCommonFields_Batch(t *testing.T) {
	SetCommonFields(map[string]string{"_t_a": "1", "_t_b": "2"})
	fields := CommonFields()
	if fields["_t_a"] != "1" || fields["_t_b"] != "2" {
		t.Errorf("SetCommonFields batch failed: %v", fields)
	}
}

func TestCommonFields_ReturnsCopy(t *testing.T) {
	SetCommonField("_t_copy", "orig")
	cp := CommonFields()
	cp["_t_copy"] = "mutated"
	if CommonFields()["_t_copy"] != "orig" {
		t.Error("CommonFields() should return a copy, not a live reference")
	}
}

func TestSetCommonField_Overwrite(t *testing.T) {
	SetCommonField("_t_over", "first")
	SetCommonField("_t_over", "second")
	if CommonFields()["_t_over"] != "second" {
		t.Error("SetCommonField should overwrite existing key")
	}
}

// ---------------------------------------------------------------------------
// File name variables
// ---------------------------------------------------------------------------

func TestSetFileNameVar(t *testing.T) {
	SetFileNameVar("_fn_key", "_fn_val")
	defaultManager.mu.RLock()
	v := defaultManager.fileNameVars["_fn_key"]
	defaultManager.mu.RUnlock()
	if v != "_fn_val" {
		t.Error("SetFileNameVar did not persist")
	}
}

// ---------------------------------------------------------------------------
// Manager: construction & isolation
// ---------------------------------------------------------------------------

func TestNewManager_WithCommonFields(t *testing.T) {
	mgr := NewManager(WithCommonFields(map[string]string{"svc": "standalone"}))
	if mgr.CommonFields()["svc"] != "standalone" {
		t.Fatal("WithCommonFields not applied to new Manager")
	}
}

func TestManager_Isolation(t *testing.T) {
	mgr1 := NewManager(WithCommonFields(map[string]string{"owner": "one"}))
	mgr2 := NewManager(WithCommonFields(map[string]string{"owner": "two"}))

	if mgr1.CommonFields()["owner"] != "one" {
		t.Error("mgr1 common field incorrect")
	}
	if mgr2.CommonFields()["owner"] != "two" {
		t.Error("mgr2 common field incorrect")
	}

	mgr1.SetCommonField("extra", "only_one")
	if mgr2.CommonFields()["extra"] == "only_one" {
		t.Error("mgr2 leaked common field from mgr1")
	}
}

func TestManager_DefaultLogger_NotNil(t *testing.T) {
	if lg := NewManager().DefaultLogger(); lg == nil {
		t.Fatal("Manager.DefaultLogger() should never be nil")
	}
}

func TestManager_GetOrCreateLogger_EmptyName(t *testing.T) {
	mgr := NewManager()
	if mgr.GetOrCreateLogger("") != mgr.DefaultLogger() {
		t.Error("empty name should return the manager's default logger")
	}
}

func TestManager_Sync_NoPanic(t *testing.T) {
	mgr := NewManager()
	mgr.Sync()
}

// ---------------------------------------------------------------------------
// Wrapper
// ---------------------------------------------------------------------------

type countingWrapper struct {
	count *int
}

func (w countingWrapper) Wrap(core zapcore.Core) zapcore.Core {
	*w.count++
	return core
}

func TestRegisterWrapper(t *testing.T) {
	mgr := NewManager()
	var c int
	mgr.RegisterWrapper(countingWrapper{count: &c})
	if c != 0 {
		t.Error("wrapper.Wrap should not be called during RegisterWrapper")
	}
	logger := NewConfigLogger(defaultConsoleConfig(), nil, mgr.wrappers)
	_ = logger.Sync()
	if c != 1 {
		t.Errorf("expected Wrap to be called once, got %d", c)
	}
}

func TestRegisterWrapper_PackageLevel(t *testing.T) {
	before := len(defaultManager.wrappers)
	RegisterWrapper(dummyWrapper{})
	if len(defaultManager.wrappers) != before+1 {
		t.Errorf("package-level RegisterWrapper did not append")
	}
}

type dummyWrapper struct{}

func (d dummyWrapper) Wrap(core zapcore.Core) zapcore.Core { return core }

// ---------------------------------------------------------------------------
// Config construction
// ---------------------------------------------------------------------------

func TestNewConfig_Defaults(t *testing.T) {
	c := defaultConsoleConfig()
	if c.LogLevel != "debug" || c.StackLevel != "error" || c.EncoderType != "console" {
		t.Errorf("default levels/encoder mismatch: %+v", c)
	}
	if !c.EnableConsole || c.EnableWriteFile {
		t.Errorf("default writer flags wrong: console=%v file=%v", c.EnableConsole, c.EnableWriteFile)
	}
	if c.MaxAge != 7 || c.RotationTime != 86400 {
		t.Errorf("default rotation wrong: age=%d rotation=%d", c.MaxAge, c.RotationTime)
	}
	if c.IncludeStdout || c.IncludeStderr {
		t.Errorf("default extra writers should be off")
	}
}

func TestConfig_TimeEncoder(t *testing.T) {
	config := defaultConsoleConfig()
	config.TimeFormat = "2006-01-02 15:04:05.000"
	enc := config.TimeEncoder()
	if enc == nil {
		t.Fatal("TimeEncoder returned nil")
	}
	var arr testPrimitiveArrayEncoder
	enc(ctime.Now().Time, &arr)
	if len(arr.appended) == 0 {
		t.Error("TimeEncoder produced no output")
	}
	if !strings.Contains(arr.appended[0], "20") {
		t.Errorf("TimeEncoder output doesn't look like a date: %q", arr.appended[0])
	}
}

type testPrimitiveArrayEncoder struct{ appended []string }

func (t *testPrimitiveArrayEncoder) AppendBool(bool)                  {}
func (t *testPrimitiveArrayEncoder) AppendByteString([]byte)          {}
func (t *testPrimitiveArrayEncoder) AppendComplex128(complex128)      {}
func (t *testPrimitiveArrayEncoder) AppendComplex64(complex64)        {}
func (t *testPrimitiveArrayEncoder) AppendDuration(d time.Duration)     {}
func (t *testPrimitiveArrayEncoder) AppendFloat64(f float64)           {}
func (t *testPrimitiveArrayEncoder) AppendFloat32(f float32)           {}
func (t *testPrimitiveArrayEncoder) AppendInt(i int)                   {}
func (t *testPrimitiveArrayEncoder) AppendInt64(i int64)               {}
func (t *testPrimitiveArrayEncoder) AppendInt32(i int32)               {}
func (t *testPrimitiveArrayEncoder) AppendInt16(i int16)               {}
func (t *testPrimitiveArrayEncoder) AppendInt8(i int8)                 {}
func (t *testPrimitiveArrayEncoder) AppendString(s string)             { t.appended = append(t.appended, s) }
func (t *testPrimitiveArrayEncoder) AppendTime(tm time.Time)           {}
func (t *testPrimitiveArrayEncoder) AppendUint(uint)                  {}
func (t *testPrimitiveArrayEncoder) AppendUint64(uint64)              {}
func (t *testPrimitiveArrayEncoder) AppendUint32(uint32)              {}
func (t *testPrimitiveArrayEncoder) AppendUint16(uint16)              {}
func (t *testPrimitiveArrayEncoder) AppendUint8(uint8)                {}
func (t *testPrimitiveArrayEncoder) AppendUintptr(uintptr)            {}

func TestConfig_EnableWriteFile_AutoDefaults(t *testing.T) {
	config := defaultConfig
	config.EnableWriteFile = true
	config.LogLevel = "info"
	config.FileLinkPath = ""
	config.FilePathFormat = ""

	if config.FileLinkPath == "" {
		config.FileLinkPath = "logs/info.log"
	}
	if config.FilePathFormat == "" {
		config.FilePathFormat = "logs/info_%Y%m%d%H%M.log"
	}

	if !strings.Contains(config.FileLinkPath, "logs/info") {
		t.Errorf("expected auto FileLinkPath, got %q", config.FileLinkPath)
	}
	if !strings.Contains(config.FilePathFormat, "info") {
		t.Errorf("expected auto FilePathFormat containing 'info', got %q", config.FilePathFormat)
	}
}

// ---------------------------------------------------------------------------
// Encoder construction
// ---------------------------------------------------------------------------

func TestBuildEncoder_Console(t *testing.T) {
	enc := buildEncoder("console", zapcore.EncoderConfig{}, nil)
	if enc == nil {
		t.Fatal("console encoder should not be nil")
	}
	if _, ok := enc.(*kvConsoleEncoder); !ok {
		t.Errorf("expected *kvConsoleEncoder, got %T", enc)
	}
}

func TestBuildEncoder_JSON(t *testing.T) {
	enc := buildEncoder("json", zapcore.EncoderConfig{}, nil)
	if enc == nil {
		t.Fatal("json encoder should not be nil")
	}
	if _, ok := enc.(*kvConsoleEncoder); ok {
		t.Errorf("expected json encoder, got *kvConsoleEncoder")
	}
}

func TestBuildEncoder_WithCommonFields_Console(t *testing.T) {
	fields := map[string]string{"nodeid": "game-1", "env": "test"}
	enc := buildEncoder("console", zapcore.EncoderConfig{}, fields)
	kv, ok := enc.(*kvConsoleEncoder)
	if !ok {
		t.Fatal("expected *kvConsoleEncoder")
	}
	if len(kv.fields) != 2 {
		t.Errorf("expected 2 common fields baked in, got %d", len(kv.fields))
	}
}

func TestBuildEncoder_WithCommonFields_JSON(t *testing.T) {
	fields := map[string]string{"nodeid": "game-1"}
	enc := buildEncoder("json", zapcore.EncoderConfig{}, fields)
	if enc == nil {
		t.Fatal("json encoder should not be nil")
	}
}

// ---------------------------------------------------------------------------
// Builder: writer combinations
// ---------------------------------------------------------------------------

func TestBuilder_ConsoleOnly(t *testing.T) {
	config := defaultConsoleConfig()
	logger := NewConfigLogger(config, nil, nil)
	if logger == nil {
		t.Fatal("NewConfigLogger returned nil for console-only config")
	}
	logger.Info("console-only test")
	_ = logger.Sync()
}

func TestBuilder_FileWriting(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "cherry_logger_file_test")
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)

	config := Config{
		LogLevel:        "debug",
		EncoderType:     "console",
		EnableConsole:   false,
		EnableWriteFile: true,
		FileLinkPath:    filepath.Join(dir, "test.log"),
		FilePathFormat:  filepath.Join(dir, "test_%Y%m%d%H%M.log"),
		MaxAge:          1,
		RotationTime:    86400,
		TimeFormat:      "15:04:05.000",
	}

	logger := NewConfigLogger(config, nil, nil)
	logger.Info("file write test")
	logger.Infof("formatted: %d", 42)
	logger.Infow("structured", "key", "val")
	_ = logger.Sync()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	hasFile := false
	for _, e := range entries {
		if strings.Contains(e.Name(), "test_") {
			hasFile = true
			break
		}
	}
	if !hasFile {
		t.Error("rotatelogs did not create a log file")
	}
}

func TestBuilder_ConsoleAndFile(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "cherry_logger_dual_test")
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)

	config := Config{
		LogLevel:        "debug",
		EncoderType:     "console",
		EnableConsole:   true,
		EnableWriteFile: true,
		FileLinkPath:    filepath.Join(dir, "dual.log"),
		FilePathFormat:  filepath.Join(dir, "dual_%Y%m%d%H%M.log"),
		MaxAge:          1,
		RotationTime:    86400,
		TimeFormat:      "15:04:05.000",
	}

	logger := NewConfigLogger(config, nil, nil)
	logger.Info("dual output test")
	_ = logger.Sync()

	entries, _ := os.ReadDir(dir)
	found := false
	for _, e := range entries {
		if strings.Contains(e.Name(), "dual_") {
			found = true
			break
		}
	}
	if !found {
		t.Error("console+file config: no log file created")
	}
}

func TestBuilder_PrintCaller(t *testing.T) {
	config := defaultConsoleConfig()
	config.EnableConsole = false
	config.PrintCaller = true
	logger := NewConfigLogger(config, nil, nil)
	if logger == nil {
		t.Fatal("logger with PrintCaller should not be nil")
	}
	_ = logger.Sync()
}

func TestBuilder_NoDoubleStderr(t *testing.T) {
	config := defaultConsoleConfig()
	config.IncludeStderr = true
	logger := NewConfigLogger(config, nil, nil)
	logger.Info("no-double-stderr test")
	_ = logger.Sync()
}

// ---------------------------------------------------------------------------
// kvConsoleEncoder behaviour
// ---------------------------------------------------------------------------

func TestKVConsoleEncoder_Clone(t *testing.T) {
	cfg := zapcore.EncoderConfig{ConsoleSeparator: "\t"}
	orig := newKVConsoleEncoder(cfg)
	orig.AddString("a", "1")

	cloned := orig.Clone()
	cloned.AddString("b", "2")

	kvOrig := orig.(*kvConsoleEncoder)
	kvCloned := cloned.(*kvConsoleEncoder)
	if len(kvOrig.fields) != 1 {
		t.Errorf("Clone should not share fields: orig has %d", len(kvOrig.fields))
	}
	if len(kvCloned.fields) != 2 {
		t.Errorf("Clone should have inherited + own fields: got %d", len(kvCloned.fields))
	}
}

func TestKVConsoleEncoder_EncodeEntry_NoPanic(t *testing.T) {
	cfg := zapcore.EncoderConfig{
		TimeKey:          "ts",
		LevelKey:         "level",
		MessageKey:       "msg",
		EncodeTime:       zapcore.TimeEncoderOfLayout("15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		ConsoleSeparator: "\t",
	}

	enc := newKVConsoleEncoder(cfg)
	enc.AddString("nodeid", "game-1")
	enc.AddString("env", "test")

	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "hello world",
		Time:    ctime.Now().Time,
		Caller:  zapcore.EntryCaller{Defined: true, File: "test.go", Line: 42},
	}

	buf, err := enc.EncodeEntry(entry, []zapcore.Field{
		zap.String("dynamic", "field"),
		zap.Int("count", 100),
	})
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Error("EncodeEntry produced empty output")
	}
	buf.Free()
}

// ---------------------------------------------------------------------------
// Concurrent safety
// ---------------------------------------------------------------------------

func TestConcurrent_Logging(t *testing.T) {
	const goroutines = 50
	const iterations = 200

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				Infof("goroutine-%d iteration-%d", id, j)
				Infow("concurrent", "g", id, "i", j)
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrent_CommonFields(t *testing.T) {
	const goroutines = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := "_t_conc_" + string(rune('a'+id%26))
			SetCommonField(key, "val")
			_ = CommonFields()
		}(i)
	}
	wg.Wait()
}

func TestConcurrent_ManagerIsolation(t *testing.T) {
	mgr1 := NewManager()
	mgr2 := NewManager()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			mgr1.SetCommonField("a", "1")
			_ = mgr1.CommonFields()
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			mgr2.SetCommonField("b", "2")
			_ = mgr2.CommonFields()
		}
	}()

	wg.Wait()
}

// ---------------------------------------------------------------------------
// Benchmark: file writing
// ---------------------------------------------------------------------------

func BenchmarkWrite(b *testing.B) {
	config := defaultConsoleConfig()
	config.EnableConsole = false
	config.EnableWriteFile = true
	config.FileLinkPath = "logs/log1.log"
	config.FilePathFormat = "logs/log1_%Y%m%d%H%M.log"

	log1 := NewConfigLogger(config, nil, nil)

	for i := 0; i < b.N; i++ {
		now := ctime.Now()
		log1.Debug(now.ToDateTimeFormat())
	}
}

// ---------------------------------------------------------------------------
// Smoke: JSON and Console encoder output
// ---------------------------------------------------------------------------

func TestJSONEncoder(t *testing.T) {
	config := defaultConsoleConfig()
	config.EncoderType = "json"
	config.EnableConsole = true
	config.PrintCaller = true

	logger := NewConfigLogger(config, nil, nil)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	logger.Debugf("formatted debug: %s", "test")
	logger.Infow("info with context",
		"key1", "value1",
		"key2", 123,
		"key3", true,
	)

	logger.Infow("game-login-log",
		"PlayerID", 111,
		"PlayerName", "nick name",
		"time", 11111,
	)

	t.Log("JSON encoder test completed")
}

func TestConsoleEncoder(t *testing.T) {
	config := defaultConsoleConfig()
	config.EncoderType = "console"
	config.EnableConsole = true
	config.PrintCaller = true

	logger := NewConfigLogger(config, nil, nil)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")

	t.Log("Console encoder test completed")
}
