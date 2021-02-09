package main

import (
	"github.com/cherry-game/cherry/logger"
)

func main() {

	//logger1 := cherryLogger.NewLogger("test_handler")
	//
	//for i := 0; i < 10; i++ {
	//	logger1.Debugw("failed to fetch URL." + strconv.Itoa(i),
	//		"url", "http://example.com",
	//		"name", "url name",
	//	)
	//}

	logger := cherryLogger.NewConsoleLogger(&cherryLogger.Config{
		Level:           "debug",
		StackLevel:      "error",
		EnableWriteFile: false,
		EnableConsole:   true,
		FilePath:        "",
		MaxSize:         0,
		MaxAge:          0,
		MaxBackups:      0,
		Compress:        false,
		TimeFormat:      "",
		PrintTime:       false,
		PrintLevel:      false,
		PrintCaller:     false,
	})

	logger.Info("111111111111111111111111111111")

	cherryLogger.Debugf("aaaaaaaaaaaaaa %s", "aaaaa args.......")

	cherryLogger.Infow("failed to fetch URL.", "url", "http://example.com")

	cherryLogger.Infow("failed to fetch URL.",
		"url", "http://example.com",
		"name", "url name",
	)

	cherryLogger.Warnw("failed to fetch URL.",
		"url", "http://example.com",
		"name", "url name",
	)

	cherryLogger.Errorw("failed to fetch URL.",
		"url", "http://example.com",
		"name", "url name",
	)

	cherryLogger.Fatal("fdsfdfs")

}
