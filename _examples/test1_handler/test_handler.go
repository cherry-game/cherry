package main

import (
	"context"
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/session"
	"google.golang.org/protobuf/proto"
)

func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

type TestHandler struct {
	cherryHandler.Handler
}

func (t *TestHandler) OnInit() {
	t.AddEvent("testEventName", t.testEvent)

	t.AddLocals(
		t.testMethod1,
		t.testMethod2,
	)

	t.AddLocal("testLocalMethod", t.testLocalMethod)
	t.AddRemote("testRemoteMethod", t.testRemoteMethod)
}

func (t *TestHandler) testMethod1(_ context.Context, _ *cherrySession.Session, _ *cherryMessage.Message) {
	cherryLogger.Debug("execute test_handler.go in testMethod1.")
}

func (t *TestHandler) testMethod2(_ context.Context, session *cherrySession.Session, message *cherryMessage.Message) {
	cherryLogger.Debug(session, message)
}

func (t *TestHandler) testLocalMethod(_ context.Context, _ *cherrySession.Session, _ *cherryMessage.Message) {
}

func (t *TestHandler) testRemoteMethod(session *cherrySession.Session, msg proto.Message) {
	cherryLogger.Debug(session, msg)
}

func (t *TestHandler) testEvent(e cherryFacade.IEvent) {
	if e != nil {
		event, ok := e.(*TestEvent)
		if !ok {
			return
		}
		cherryLogger.Debugf("execute event data. value=%d", event.Abc)
	} else {
		//cherryLogger.Debug("execute event data is nil")
	}
}

func (t *TestHandler) testTrigger() {
	cherryLogger.Debug("test trigger " + t.Name())
}
