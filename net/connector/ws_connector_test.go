package cherryConnector

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"sync"
	"testing"
)

// websocket client http://www.websocket-test.com/
func TestNewWSConnector(t *testing.T) {

	wg := &sync.WaitGroup{}
	wg.Add(1)

	connector := NewWS(":9071")

	connector.OnConnectListener(func(conn cfacade.INetConn) {
		clog.Infof("new net.INetConn = %s", conn.RemoteAddr())
	})

	connector.OnStart()

	wg.Wait()
}
