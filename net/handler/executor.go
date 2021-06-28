package cherryHandler

import (
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryAgent "github.com/cherry-game/cherry/net/agent"
	message "github.com/cherry-game/cherry/net/message"
	"reflect"
)

type (
	IExecutor interface {
		Invoke()
	}

	MessageExecutor struct {
		Agent         *cherryAgent.Agent
		Msg           *message.Message
		HandlerFn     *facade.HandlerFn
		BeforeFilters []FilterFn
		AfterFilters  []FilterFn
	}

	EventExecutor struct {
		Event   facade.IEvent
		EventFn []facade.EventFn
	}

	UserRPCExecutor struct {
		HandlerFn *facade.HandlerFn
	}

	SysRPCExecutor struct {
		HandlerFn *facade.HandlerFn
	}
)

func (m *MessageExecutor) Invoke() {
	for _, filter := range m.BeforeFilters {
		if !filter(m) {
			break
		}
	}

	if len(m.HandlerFn.InArgs) != 2 {
		cherryLogger.Warnf("[Route = %v] method in args error", m.Msg.Route)
		return
	}

	in2 := m.HandlerFn.InArgs[1]

	var val interface{}
	val = reflect.New(in2.Elem()).Interface()

	err := m.Agent.Serializer.Unmarshal(m.Msg.Data, val)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	params := make([]reflect.Value, 2)
	params[0] = reflect.ValueOf(m.Agent.Session)
	params[1] = reflect.ValueOf(val)

	result := m.HandlerFn.Value.Call(params)
	if len(result) > 0 {
		if err := result[0].Interface(); err != nil {
			cherryLogger.Warn(err)
			return
		}
	}

	for _, filter := range m.AfterFilters {
		if !filter(m) {
			break
		}
	}
}

func (e *EventExecutor) Invoke() {
	for _, fn := range e.EventFn {
		fn(e.Event)
	}
}

func (u *UserRPCExecutor) Invoke() {

}

func (u *SysRPCExecutor) Invoke() {

}
