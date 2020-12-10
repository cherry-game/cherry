package main

import (
	"github.com/cherry-game/cherry/logger"
)

func main() {
	cherryLogger.Debugf("aaaaaaaaaaaaaa %s", "aaaaa args.......")

	cherryLogger.Infow("failed to fetch URL.", "url", "http://example.com")

	cherryLogger.Infow("failed to fetch URL.",
		"url", "http://example.com",
		"name", "url name",
	)
}
