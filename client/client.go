package cherryClient

import (
	"github.com/cherry-game/cherry/net/message"
)

type IClient interface {
	ConnectTo(addr string) error
	ConnectToTLS(addr string, skipVerify bool) error
	Disconnect()
	SendNotify(route string, data []byte) error
	SendRequest(route string, data []byte) (uint64, error)
	ConnectedStatus() bool
	MsgChannel() chan *cherryMessage.Message
}
