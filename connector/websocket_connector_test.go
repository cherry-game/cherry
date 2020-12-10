package cherryConnector

import (
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/session"
	"net"
	"testing"
)

// websocket client http://www.websocket-test.com/
func TestNewWSConnector(t *testing.T) {
	connector := NewWSConnector(":9071")

	connector.OnConnect(func(conn net.Conn) {
		s := cherrySession.NewSession(cherrySession.NextSessionId(), conn, nil, nil)
		cherryLogger.Infof("new session sid = %s, address = %s", s.SID(), s.Conn().RemoteAddr())

		s.OnMessage(func(bytes []byte) (err error) {
			text := string(bytes)
			cherryLogger.Info(text)

			if len(text) == 1 && text[0] == 99 {
				s.Closed()
			}

			return nil
		})

		s.Start()
	})

	connector.Start()
}
