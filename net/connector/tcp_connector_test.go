package cherryConnector

import (
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"testing"
)

func TestNewTCPConnector(t *testing.T) {
	connector := NewTCP(":9071")

	connector.OnConnect(func(conn facade.INetConn) {
		cherryLogger.Infof("new net.INetConn = %s", conn.RemoteAddr())

		//session := cherrySession.NewSession(cherrySession.NextSID(), "", conn, nil)
		//
		//session.OnMessage(func(bytes []byte) (err error) {
		//	cherryLogger.Infof("session id=%d bytes=[%s]", session.SID(), bytes)
		//
		//	session.Send(bytes)
		//
		//	if len(bytes) == 1 && bytes[0] == 3 {
		//		session.Close()
		//	}
		//
		//	return nil
		//})
		//
		//session.Run()
	})

	connector.OnStart()
}
