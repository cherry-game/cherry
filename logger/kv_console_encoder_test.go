package cherryLogger

import (
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

func BenchmarkKVConsoleEncoder_CommonFields(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		TimeKey:          "ts",
		LevelKey:         "level",
		CallerKey:        "caller",
		MessageKey:       "msg",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	encoder.AddString("nodeid", "game-1")
	encoder.AddString("nodetype", "game")
	encoder.AddString("env", "dev")

	entry := zapcore.Entry{
		Time:    time.Now(),
		Level:   zapcore.InfoLevel,
		Caller:  zapcore.EntryCaller{Defined: true, File: "test.go", Line: 100},
		Message: "test message",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, _ := encoder.EncodeEntry(entry, nil)
		buf.Free()
	}
}

func BenchmarkKVConsoleEncoder_DynamicFields(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		TimeKey:          "ts",
		LevelKey:         "level",
		MessageKey:       "msg",
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	encoder.AddString("nodeid", "game-1")

	entry := zapcore.Entry{
		Time:    time.Now(),
		Level:   zapcore.InfoLevel,
		Message: "test message",
	}

	fields := []zapcore.Field{
		zap.String("key1", "value1"),
		zap.String("key2", "value2"),
		zap.Int("count", 100),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, _ := encoder.EncodeEntry(entry, fields)
		buf.Free()
	}
}

func BenchmarkKVConsoleEncoder_StringFields(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	for i := 0; i < 10; i++ {
		encoder.AddString("field"+string(rune(i)), "value"+string(rune(i)))
	}

	entry := zapcore.Entry{
		Time:    time.Now(),
		Level:   zapcore.InfoLevel,
		Message: "test message",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, _ := encoder.EncodeEntry(entry, nil)
		buf.Free()
	}
}

func BenchmarkKVConsoleEncoder_IntFields(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	for i := 0; i < 10; i++ {
		encoder.AddInt("field"+string(rune(i)), i*100)
	}

	entry := zapcore.Entry{
		Time:    time.Now(),
		Level:   zapcore.InfoLevel,
		Message: "test message",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, _ := encoder.EncodeEntry(entry, nil)
		buf.Free()
	}
}

func BenchmarkKVConsoleEncoder_FloatFields(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	for i := 0; i < 10; i++ {
		encoder.AddFloat64("field"+string(rune(i)), float64(i)*1.5)
	}

	entry := zapcore.Entry{
		Time:    time.Now(),
		Level:   zapcore.InfoLevel,
		Message: "test message",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, _ := encoder.EncodeEntry(entry, nil)
		buf.Free()
	}
}

func BenchmarkKVConsoleEncoder_MixedFields(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	encoder.AddString("nodeid", "game-1")
	encoder.AddString("nodetype", "game")
	encoder.AddInt("pid", 12345)
	encoder.AddBool("debug", true)

	entry := zapcore.Entry{
		Time:    time.Now(),
		Level:   zapcore.InfoLevel,
		Message: "test message",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, _ := encoder.EncodeEntry(entry, nil)
		buf.Free()
	}
}

func BenchmarkKVConsoleEncoder_Clone(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	encoder.AddString("nodeid", "game-1")
	encoder.AddString("nodetype", "game")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.Clone()
	}
}

func BenchmarkKVConsoleEncoder_vs_StandardConsole(b *testing.B) {
	b.Run("KVConsole", func(b *testing.B) {
		cfg := zapcore.EncoderConfig{
			EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
			EncodeLevel:      zapcore.CapitalLevelEncoder,
			ConsoleSeparator: "\t",
		}

		encoder := newKVConsoleEncoder(cfg)
		encoder.AddString("nodeid", "game-1")
		encoder.AddString("nodetype", "game")

		entry := zapcore.Entry{
			Time:    time.Now(),
			Level:   zapcore.InfoLevel,
			Message: "test message",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf, _ := encoder.EncodeEntry(entry, nil)
			buf.Free()
		}
	})

	b.Run("StandardConsole", func(b *testing.B) {
		cfg := zapcore.EncoderConfig{
			EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
			EncodeLevel:      zapcore.CapitalLevelEncoder,
			ConsoleSeparator: "\t",
		}

		encoder := zapcore.NewConsoleEncoder(cfg)

		entry := zapcore.Entry{
			Time:    time.Now(),
			Level:   zapcore.InfoLevel,
			Message: "test message",
		}

		fields := []zapcore.Field{
			zap.String("nodeid", "game-1"),
			zap.String("nodetype", "game"),
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			encoder.AddString("nodeid", "game-1")
			encoder.AddString("nodetype", "game")
			buf, _ := encoder.EncodeEntry(entry, fields)
			buf.Free()
		}
	})
}

func BenchmarkWriteKVFields(b *testing.B) {
	buf := buffer.NewPool().Get()

	fields := []zapcore.Field{
		zap.String("nodeid", "game-1"),
		zap.String("nodetype", "game"),
		zap.Int("count", 100),
		zap.Bool("debug", true),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeKVFields(buf, fields)
		buf.Reset()
	}
}

func BenchmarkWriteKVFields_StringOnly(b *testing.B) {
	buf := buffer.NewPool().Get()

	fields := []zapcore.Field{
		zap.String("nodeid", "game-1"),
		zap.String("nodetype", "game"),
		zap.String("env", "dev"),
		zap.String("version", "1.0.0"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeKVFields(buf, fields)
		buf.Reset()
	}
}

func BenchmarkWriteKVFields_Large(b *testing.B) {
	buf := buffer.NewPool().Get()

	fields := make([]zapcore.Field, 50)
	for i := 0; i < 50; i++ {
		fields[i] = zap.String("field"+string(rune(i%10)), "value"+string(rune(i%10)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeKVFields(buf, fields)
		buf.Reset()
	}
}
