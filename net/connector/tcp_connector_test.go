package cherryConnector

import (
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"sync"
	"testing"
)

func TestNewTCPConnector(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	connector := NewTCP(":9071")

	connector.OnConnectListener(func(conn facade.INetConn) {
		cherryLogger.Infof("new net.INetConn = %s", conn.RemoteAddr())
	})

	connector.OnStart()

	wg.Wait()
}
