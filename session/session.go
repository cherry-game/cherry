package cherrySession

import (
	"fmt"
	"github.com/phantacix/cherry/interfaces"
	"github.com/phantacix/cherry/logger"
	"github.com/phantacix/cherry/net/pomelo_packet"
	"github.com/phantacix/cherry/utils"
	"net"
	"sync/atomic"
	"time"
)

var (
	IllegalUID = cherryUtils.Error("illegal uid")
)

type Session struct {
	Settings
	status            int
	conn              net.Conn
	running           bool
	sid               cherryInterfaces.SID            // session id
	uid               cherryInterfaces.UID            // user unique id
	frontendId        cherryInterfaces.FrontendId     // frontend node id
	net               cherryInterfaces.INetworkEntity // network opts
	sessionComponent  *SessionComponent               // session SessionComponent
	lastTime          int64                           // last update time
	sendChan          chan []byte
	onCloseListener   []cherryInterfaces.SessionListener
	onErrorListener   []cherryInterfaces.SessionListener
	onMessageListener cherryInterfaces.MessageListener
}

func NewSession(sid cherryInterfaces.SID, conn net.Conn, net cherryInterfaces.INetworkEntity, sessionComponent *SessionComponent) *Session {
	session := &Session{
		Settings: Settings{
			data: make(map[string]interface{}),
		},
		status:           INIT,
		conn:             conn,
		running:          false,
		sid:              sid,
		uid:              0,
		frontendId:       "",
		net:              net,
		sessionComponent: sessionComponent,
		lastTime:         time.Now().Unix(),
		sendChan:         make(chan []byte),
	}

	session.onCloseListener = append(session.onCloseListener, func(session cherryInterfaces.ISession) {
		cherryLogger.Infof("on closed. %s", session)
	})

	session.onErrorListener = append(session.onErrorListener, func(session cherryInterfaces.ISession) {
		cherryLogger.Infof("on error. %s", session)
	})

	return session
}

func (s *Session) SID() cherryInterfaces.SID {
	return s.sid
}

func (s *Session) UID() cherryInterfaces.UID {
	return s.uid
}

func (s *Session) FrontendId() cherryInterfaces.FrontendId {
	return s.frontendId
}

func (s *Session) SetStatus(status int) {
	s.status = status
}

func (s *Session) Status() int {
	return s.status
}

func (s *Session) Net() cherryInterfaces.INetworkEntity {
	return s.net
}

func (s *Session) Bind(uid cherryInterfaces.UID) error {
	if uid < 1 {
		return IllegalUID
	}

	atomic.StoreInt64(&s.uid, uid)
	return nil
}

func (s *Session) Conn() net.Conn {
	return s.conn
}

func (s *Session) OnClose(listener cherryInterfaces.SessionListener) {
	if listener != nil {
		s.onCloseListener = append(s.onCloseListener, listener)
	}
}

func (s *Session) OnError(listener cherryInterfaces.SessionListener) {
	if listener != nil {
		s.onErrorListener = append(s.onErrorListener, listener)
	}
}

func (s *Session) OnMessage(listener cherryInterfaces.MessageListener) {
	if listener != nil {
		s.onMessageListener = listener
	}
}

func (s *Session) Send(msg []byte) error {
	if !s.running {
		return nil
	}

	s.sendChan <- msg
	return nil
}

func (s *Session) SendBatch(batchMsg ...[]byte) {
	for _, msg := range batchMsg {
		s.Send(msg)
	}
}

func (s *Session) Start() {
	s.running = true

	// read goroutine
	go s.readPackets(2048)

	for s.running {
		select {
		case msg := <-s.sendChan:
			_, err := s.conn.Write(msg)
			if err != nil {
				s.Closed()
			}
		}
	}
}

// readPackets read connection data stream
func (s *Session) readPackets(readSize int) {
	if s.onMessageListener == nil {
		panic("onMessageListener() not set.")
	}

	defer func() {
		for _, listener := range s.onCloseListener {
			listener(s)
		}

		//close connection
		err := s.conn.Close()
		if err != nil {
			cherryLogger.Error(err)
		}
	}()

	buf := make([]byte, readSize)

	for s.running {
		n, err := s.conn.Read(buf)
		if err != nil {
			cherryLogger.Warnf("read message error: %s, socket will be closed immediately", err.Error())
			for _, listener := range s.onErrorListener {
				listener(s)
			}
			s.running = false
			return
		}

		if n < 1 {
			continue
		}

		// (warning): decoder use slice for performance, packet data should be copy before next PacketDecode
		err = s.onMessageListener(buf[:n])
		if err != nil {
			cherryLogger.Warn(err)
			return
		}
	}
}

func (s *Session) Closed() {
	s.status = CLOSED
	s.running = false

	if s.sessionComponent != nil {
		s.sessionComponent.Remove(s.sid)
	}
}

func (s *Session) String() string {
	return fmt.Sprintf("sid = %d, uid = %d, status=%s, address = %s, running = %v",
		s.sid,
		s.uid,
		cherryPomeloPacket.SessionStatus[s.status],
		s.conn.RemoteAddr().String(),
		s.running)
}
