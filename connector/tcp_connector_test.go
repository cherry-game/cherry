package cherryConnector

import (
	"github.com/phantacix/cherry/logger"
	"github.com/phantacix/cherry/session"
	"net"
	"testing"
)

func TestNewTCPConnector(t *testing.T) {
	cherryLogger.DefaultSet()

	connector := NewTCPConnector(":9071")

	connector.OnConnect(func(conn net.Conn) {
		cherryLogger.Infof("new net.Conn = %s", conn.RemoteAddr())

		session := cherrySession.NewSession(cherrySession.NextSessionId(), conn, nil, nil)

		session.OnMessage(func(bytes []byte) (err error) {
			cherryLogger.Infof("session id=%d bytes=[%s]", session.SID(), bytes)

			session.Send(bytes)

			if len(bytes) == 1 && bytes[0] == 3 {
				session.Closed()
			}

			return nil
		})

		session.Start()
	})

	connector.Start()
}
