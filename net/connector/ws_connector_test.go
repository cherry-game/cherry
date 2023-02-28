package cherryConnector

import (
	"fmt"
	clog "github.com/cherry-game/cherry/logger"
	"net"
	"sync"
	"testing"
)

// websocket client http://www.websocket-test.com/
func TestNewWSConnector(t *testing.T) {

	wg := &sync.WaitGroup{}
	wg.Add(1)

	ws := NewWS(":9071")
	ws.OnConnect(func(conn net.Conn) {
		clog.Infof("new net.Conn = %s", conn.RemoteAddr())
		go func() {
			for {
				buf := make([]byte, 2048)
				for {
					n, err := conn.Read(buf)
					if err != nil {
						return
					}
					fmt.Println(buf[:n])
				}
			}
		}()
	})
	ws.Start()

	wg.Wait()
}
