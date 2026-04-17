package cherryLogger

import (
	"bytes"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func BenchmarkLogger_Info(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		TimeKey:          "ts",
		LevelKey:         "level",
		CallerKey:        "caller",
		MessageKey:       "msg",
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	encoder.AddString("nodeid", "game-1")
	encoder.AddString("nodetype", "game")

	ws := zapcore.AddSync(&bytes.Buffer{})
	core := zapcore.NewCore(encoder, ws, zapcore.InfoLevel)
	logger := zap.New(core).Sugar()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("test message")
	}
}

func BenchmarkLogger_Infof(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		TimeKey:          "ts",
		LevelKey:         "level",
		CallerKey:        "caller",
		MessageKey:       "msg",
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	encoder.AddString("nodeid", "game-1")
	encoder.AddString("nodetype", "game")

	ws := zapcore.AddSync(&bytes.Buffer{})
	core := zapcore.NewCore(encoder, ws, zapcore.InfoLevel)
	logger := zap.New(core).Sugar()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Infof("player %d login", i)
	}
}

func BenchmarkLogger_Info_WithFields(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		TimeKey:          "ts",
		LevelKey:         "level",
		CallerKey:        "caller",
		MessageKey:       "msg",
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	encoder.AddString("nodeid", "game-1")
	encoder.AddString("nodetype", "game")

	ws := zapcore.AddSync(&bytes.Buffer{})
	core := zapcore.NewCore(encoder, ws, zapcore.InfoLevel)
	logger := zap.New(core).Sugar()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Infow("test message", "playerId", 10001, "action", "login")
	}
}

func BenchmarkLogger_Debug_Disabled(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	encoder.AddString("nodeid", "game-1")

	ws := zapcore.AddSync(&bytes.Buffer{})
	core := zapcore.NewCore(encoder, ws, zapcore.InfoLevel)
	logger := zap.New(core).Sugar()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("test message")
	}
}

func BenchmarkLogger_Info_vs_Standard(b *testing.B) {
	b.Run("KVConsole", func(b *testing.B) {
		cfg := zapcore.EncoderConfig{
			EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
			EncodeLevel:      zapcore.CapitalLevelEncoder,
			ConsoleSeparator: "\t",
		}

		encoder := newKVConsoleEncoder(cfg)
		encoder.AddString("nodeid", "game-1")
		encoder.AddString("nodetype", "game")

		ws := zapcore.AddSync(&bytes.Buffer{})
		core := zapcore.NewCore(encoder, ws, zapcore.InfoLevel)
		logger := zap.New(core).Sugar()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("test message")
		}
	})

	b.Run("StandardConsole", func(b *testing.B) {
		cfg := zapcore.EncoderConfig{
			EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
			EncodeLevel:      zapcore.CapitalLevelEncoder,
			ConsoleSeparator: "\t",
		}

		encoder := zapcore.NewConsoleEncoder(cfg)
		ws := zapcore.AddSync(&bytes.Buffer{})
		core := zapcore.NewCore(encoder, ws, zapcore.InfoLevel)
		logger := zap.New(core).Sugar()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("test message", zap.String("nodeid", "game-1"), zap.String("nodetype", "game"))
		}
	})

	b.Run("StandardJSON", func(b *testing.B) {
		cfg := zapcore.EncoderConfig{
			EncodeTime:  zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
			EncodeLevel: zapcore.CapitalLevelEncoder,
		}

		encoder := zapcore.NewJSONEncoder(cfg)
		encoder.AddString("nodeid", "game-1")
		encoder.AddString("nodetype", "game")

		ws := zapcore.AddSync(&bytes.Buffer{})
		core := zapcore.NewCore(encoder, ws, zapcore.InfoLevel)
		logger := zap.New(core).Sugar()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("test message")
		}
	})
}

func BenchmarkLogger_HighFrequency(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	encoder.AddString("nodeid", "game-1")
	encoder.AddString("nodetype", "game")

	ws := zapcore.AddSync(&bytes.Buffer{})
	core := zapcore.NewCore(encoder, ws, zapcore.InfoLevel)
	logger := zap.New(core).Sugar()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("player action")
		logger.Infof("player %d login", i)
		logger.Infow("battle result", "playerId", 10001, "score", 95)
	}
}

func BenchmarkLogger_RealWorld(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	encoder.AddString("nodeid", "game-1")
	encoder.AddString("nodetype", "game")
	encoder.AddInt("pid", 12345)

	ws := zapcore.AddSync(&bytes.Buffer{})
	core := zapcore.NewCore(encoder, ws, zapcore.DebugLevel)
	logger := zap.New(core, zap.AddCaller()).Sugar()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debugf("request received, playerID=%d", 10001)
		logger.Infof("player %d login success", 10001)
		logger.Infow("battle start", "playerId", 10001, "heroId", 100)
		logger.Warnw("low resources", "type", "gold", "amount", 100)
	}
}

func BenchmarkLogger_Infow_ManyFields(b *testing.B) {
	cfg := zapcore.EncoderConfig{
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		ConsoleSeparator: "\t",
	}

	encoder := newKVConsoleEncoder(cfg)
	encoder.AddString("nodeid", "game-1")

	ws := zapcore.AddSync(&bytes.Buffer{})
	core := zapcore.NewCore(encoder, ws, zapcore.InfoLevel)
	logger := zap.New(core).Sugar()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Infow("battle log",
			"playerId", 10001,
			"heroId", 100,
			"enemyId", 200,
			"damage", 1500,
			"result", "win",
			"time", 30,
		)
	}
}

func BenchmarkLogger_RealWorld_vs_Standard(b *testing.B) {
	b.Run("KVConsole", func(b *testing.B) {
		cfg := zapcore.EncoderConfig{
			EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
			EncodeLevel:      zapcore.CapitalLevelEncoder,
			EncodeCaller:     zapcore.ShortCallerEncoder,
			ConsoleSeparator: "\t",
		}

		encoder := newKVConsoleEncoder(cfg)
		encoder.AddString("nodeid", "game-1")
		encoder.AddString("nodetype", "game")
		encoder.AddInt("pid", 12345)

		ws := zapcore.AddSync(&bytes.Buffer{})
		core := zapcore.NewCore(encoder, ws, zapcore.DebugLevel)
		logger := zap.New(core, zap.AddCaller()).Sugar()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debugf("request received, playerID=%d", 10001)
			logger.Infof("player %d login success", 10001)
			logger.Infow("battle start", "playerId", 10001, "heroId", 100)
			logger.Warnw("low resources", "type", "gold", "amount", 100)
		}
	})

	b.Run("StandardConsole", func(b *testing.B) {
		cfg := zapcore.EncoderConfig{
			EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
			EncodeLevel:      zapcore.CapitalLevelEncoder,
			EncodeCaller:     zapcore.ShortCallerEncoder,
			ConsoleSeparator: "\t",
		}

		encoder := zapcore.NewConsoleEncoder(cfg)
		ws := zapcore.AddSync(&bytes.Buffer{})
		core := zapcore.NewCore(encoder, ws, zapcore.DebugLevel)
		logger := zap.New(core, zap.AddCaller()).Sugar()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debugf("request received, playerID=%d", 10001)
			logger.Infof("player %d login success", 10001)
			logger.Infow("battle start",
				zap.String("nodeid", "game-1"),
				zap.String("nodetype", "game"),
				zap.Int("pid", 12345),
				zap.Int("playerId", 10001),
				zap.Int("heroId", 100),
			)
			logger.Warnw("low resources",
				zap.String("nodeid", "game-1"),
				zap.String("nodetype", "game"),
				zap.Int("pid", 12345),
				zap.String("type", "gold"),
				zap.Int("amount", 100),
			)
		}
	})

	b.Run("StandardJSON", func(b *testing.B) {
		cfg := zapcore.EncoderConfig{
			EncodeTime:   zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		}

		encoder := zapcore.NewJSONEncoder(cfg)
		encoder.AddString("nodeid", "game-1")
		encoder.AddString("nodetype", "game")
		encoder.AddInt("pid", 12345)

		ws := zapcore.AddSync(&bytes.Buffer{})
		core := zapcore.NewCore(encoder, ws, zapcore.DebugLevel)
		logger := zap.New(core, zap.AddCaller()).Sugar()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debugf("request received, playerID=%d", 10001)
			logger.Infof("player %d login success", 10001)
			logger.Infow("battle start", "playerId", 10001, "heroId", 100)
			logger.Warnw("low resources", "type", "gold", "amount", 100)
		}
	})
}
