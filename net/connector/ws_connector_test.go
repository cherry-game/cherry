package cherryConnector

import (
	"fmt"
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

		go func() {
			for {
				msg, err := conn.GetNextMessage()
				fmt.Println(msg, err)
			}
		}()
	})

	connector.OnStart()

	wg.Wait()
}
