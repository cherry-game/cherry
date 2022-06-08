package cherryConnector

import (
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"sync"
	"testing"
)

// websocket client http://www.websocket-test.com/
func TestNewWSConnector(t *testing.T) {

	wg := &sync.WaitGroup{}
	wg.Add(1)

	connector := NewWS(":9071")

	connector.OnConnectListener(func(conn cherryFacade.INetConn) {
		cherryLogger.Infof("new net.INetConn = %s", conn.RemoteAddr())
	})

	connector.OnStart()

	wg.Wait()
}
