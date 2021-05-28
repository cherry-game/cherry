package cherryConnector

import (
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/session"
	"net"
	"testing"
)

func TestNewTCPConnector(t *testing.T) {
	connector := NewTCPConnector(":9071")

	connector.OnConnect(func(conn net.Conn) {
		cherryLogger.Infof("new net.Conn = %s", conn.RemoteAddr())

		session := cherrySession.NewSession(cherrySession.NewSessionId(), conn, nil, nil)

		session.OnMessage(func(bytes []byte) (err error) {
			cherryLogger.Infof("session id=%d bytes=[%s]", session.SID(), bytes)

			err = session.Send(bytes)
			if err != nil {
				return err
			}

			if len(bytes) == 1 && bytes[0] == 3 {
				session.Closed()
			}

			return nil
		})

		session.Start()
	})

	connector.OnStart()
}
