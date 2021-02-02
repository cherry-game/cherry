package mocks

import (
	"context"
	"github.com/cherry-game/cherry/handler"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/message"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"strconv"
)

func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

type TestHandler struct {
	cherryHandler.Handler
	logger *zap.SugaredLogger
}

func (t *TestHandler) Init() {

	t.logger = cherryLogger.NewLogger("test_handler")

	for i := 0; i < 100; i++ {
		t.logger.Debug("test handler log" + strconv.Itoa(i))
	}

	t.SetWorkerCRC32Hash(10)

	t.SetWorkerExecutor(cherryHandler.DefaultWorkerExecutor)

	t.RegisterEvent("testEventName", t.testEvent)

	t.RegisterLocals(
		t.testMethod1,
		t.testMethod2,
	)

	t.RegisterLocal("testLocalMethod", t.testMethod)
	t.RegisterRemote("testRemoteMethod", t.testRemote)
}

func (t *TestHandler) testMethod1(_ cherryInterfaces.ISession, _ *cherryMessage.Message) {
	cherryLogger.Debug("execute test_handler.go in testMethod1.")
}

func (t *TestHandler) testMethod2(session cherryInterfaces.ISession, message *cherryMessage.Message) {
	cherryLogger.Debug(session, message)
}

func (t *TestHandler) testMethod(session cherryInterfaces.ISession, message *cherryMessage.Message) {
	cherryLogger.Debug(session, message)
}

func (t *TestHandler) testRemote(ctx context.Context, msg proto.Message) {
	cherryLogger.Debug(ctx, msg)
}

func (t *TestHandler) testEvent(e cherryInterfaces.IEvent) {
	if e != nil {

		event, ok := e.(*TestEvent)
		if !ok {
			return
		}

		cherryLogger.Debug("execute event data. value=%d", event.Abc)
	} else {
		//cherryLogger.Debug("execute event data is nil")
	}
}

func (t *TestHandler) testTrigger() {
	cherryLogger.Debug("test trigger " + t.Name())
}
