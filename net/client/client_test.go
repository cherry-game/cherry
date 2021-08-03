package cherryClient

import (
	"encoding/json"
	"testing"
	"time"

	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryMessage "github.com/cherry-game/cherry/net/message"
)

func TestClient(t *testing.T) {

	c := New(NewWebSocketConn("ws://127.0.0.1:34590"), 100*time.Millisecond)
	//c := New(NewTcpConn("127.0.0.1:34590"), 100*time.Millisecond)

	c.ConnectTo()

	defer c.Disconnect()

	//发送login
	login := loginRequest{
		Nickname: "test",
	}

	msgData, _ := json.Marshal(login)

	msgID, err := c.sendMsg(cherryMessage.Request, "chat.userHandler.login", msgData)
	if nil != err {
		cherryLogger.Infof("send msgid:%d fail", msgID)
	}

	for msg := range c.MsgChannel() {

		cherryLogger.Infof("rcv msg route:%s msgid:%d msg:%s", msg.Route, msg.ID, msg.Data)
	}
}
