package cherryConnector

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"sync"
	"testing"
)

func TestNewTCPConnector(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	tcp := NewTCP(":9071")

	tcp.OnConnectListener(func(conn cfacade.INetConn) {
		clog.Infof("new net.INetConn = %s", conn.RemoteAddr())
	})

	tcp.OnStart()

	wg.Wait()
}
