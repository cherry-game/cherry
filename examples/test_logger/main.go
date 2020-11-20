package main

import (
	"github.com/phantacix/cherry/logger"
)

func main() {

	cherryLogger.DefaultSet()

	cherryLogger.Debugf("aaaaaaaaaaaaaa %s", "aaaaa args.......")

	cherryLogger.Infow("failed to fetch URL.", "url", "http://example.com")

	cherryLogger.Infow("failed to fetch URL.",
		"url", "http://example.com",
		"name", "url name",
	)
}
