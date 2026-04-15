package cherryLogger

import (
	"testing"

	ctime "github.com/cherry-game/cherry/extend/time"
)

func BenchmarkWrite(b *testing.B) {
	config := defaultConsoleConfig()
	config.EnableConsole = false
	config.EnableWriteFile = true
	config.FileLinkPath = "logs/log1.log"
	config.FilePathFormat = "logs/log1_%Y%m%d%H%M.log"

	log1 := NewConfigLogger(config)

	for i := 0; i < b.N; i++ {
		now := ctime.Now()
		log1.Debug(now.ToDateTimeFormat())
	}
}

func TestJSONEncoder(t *testing.T) {
	config := defaultConsoleConfig()
	config.EncoderType = "json"
	config.EnableConsole = true
	config.PrintCaller = true

	logger := NewConfigLogger(config)

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

	logger := NewConfigLogger(config)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")

	t.Log("Console encoder test completed")
}
