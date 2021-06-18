package cherryHandler

import (
	cherryCrypto "github.com/cherry-game/cherry/extend/crypto"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherrySession "github.com/cherry-game/cherry/net/session"
	"reflect"
)

type (
	IExecutor interface {
		HashQueue(queueNum int) int
		Invoke()
	}

	MessageExecutor struct {
		Session       *cherrySession.Session
		Msg           *cherryMessage.Message
		HandlerFn     *cherryFacade.HandlerFn
		BeforeFilters []FilterFn
		AfterFilters  []FilterFn
	}

	EventExecutor struct {
		Event   cherryFacade.IEvent
		EventFn []cherryFacade.EventFn
	}

	UserRPCExecutor struct {
		HandlerFn *cherryFacade.HandlerFn
	}

	SysRPCExecutor struct {
		HandlerFn *cherryFacade.HandlerFn
	}
)

func (m *MessageExecutor) HashQueue(queueNum int) int {
	val := m.Session.UID()
	if val == "" {
		val = string(m.Session.SID())
	}
	return cherryCrypto.CRC32(val) % queueNum
}

func (m *MessageExecutor) Invoke() {
	for _, filter := range m.BeforeFilters {
		if !filter(m) {
			break
		}
	}

	params := make([]reflect.Value, 2)
	params[0] = reflect.ValueOf(m.Session)
	params[1] = reflect.ValueOf(m.Msg)
	m.HandlerFn.Value.Call(params)

	for _, filter := range m.AfterFilters {
		if !filter(m) {
			break
		}
	}
}

func (e *EventExecutor) HashQueue(queueNum int) int {
	return cherryCrypto.CRC32(e.Event.UniqueId()) % queueNum
}

func (e *EventExecutor) Invoke() {
	for _, fn := range e.EventFn {
		fn(e.Event)
	}
}
