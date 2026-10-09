package cherryActor

import (
	"sync/atomic"
	"testing"
	"time"

	ccode "github.com/cherry-game/cherry/code"
	cfacade "github.com/cherry-game/cherry/facade"
)

// 回归：同节点 CallWait 必须在返回的 Message 被接收方 Recycle() 之后仍然拿到
// 自己的响应通道。
//
// 接收方 Actor 在另一个 goroutine 上处理完消息后会 defer Recycle()，把
// Message.ChanResult 置 nil 并放回复用池。旧实现的 select 每次都重新读取
// `message.ChanResult` 字段，于是存在一个窗口：字段已被置 nil，调用方却还没
// 进入 select，调用方就会选中 nil channel，只能等 callTimeout 超时（响应丢失）；
// 若该 Message 在此期间被池复用，调用方还会读到另一个请求的通道并收到它的响应
// （响应串用）。
//
// 两个断言：
//  1. 20000 次同节点 CallWait 全部返回 OK，不允许出现超时；
//  2. 必须用 `go test -race` 运行，竞态检测器会直接报告 message.go 的
//     `p.ChanResult = nil` 与 system.go 调用方读取该字段之间的数据竞争。
//
// 修复前：-race 稳定报 DATA RACE（非 race 模式下也约有 0.02%~0.05% 的调用
// 会超时 3 秒）。修复后：两种情况都通过。
func TestCallWaitLocalSameNodeNoRace(t *testing.T) {
	const iters = 20000

	system := NewSystem()
	system.Start(&callWaitTestApp{})
	defer system.Stop()

	receiver := &callWaitTestReceiver{}
	sender := &callWaitTestSender{iters: iters, done: make(chan struct{})}

	if _, err := system.CreateActor("receiver", receiver); err != nil {
		t.Fatalf("create receiver: %v", err)
	}
	if _, err := system.CreateActor("sender", sender); err != nil {
		t.Fatalf("create sender: %v", err)
	}

	waitWorker := func(name string, state func() State) {
		for i := 0; i < 5000; i++ {
			if state() == WorkerState {
				return
			}
			time.Sleep(time.Millisecond)
		}
		t.Fatalf("actor %s is not in WorkerState", name)
	}
	waitWorker("receiver", receiver.State)
	waitWorker("sender", sender.State)

	system.Call("callwait_test.driver", "callwait_test.sender", "run", nil)

	// 修复后本用例只需几十毫秒。出现竞态时每次失败都会白等一个
	// callTimeout(3s)，这里给足余量以便拿到精确的失败原因。
	timeout := time.After(60 * time.Second)
	select {
	case <-sender.done:
	case <-timeout:
		t.Fatalf("CallWait did not finish in time (completed receiver calls: %d)", receiver.count.Load())
	}

	if failed := sender.failed.Load(); failed != 0 {
		t.Fatalf("CallWait failed %d/%d times, last code = %d (ActorCallTimeout = %d); "+
			"receiver completed %d calls, so the replies were produced but read from a recycled Message",
			failed, iters, sender.lastCode.Load(), ccode.ActorCallTimeout, receiver.count.Load())
	}
	if got := receiver.count.Load(); got != iters {
		t.Fatalf("receiver completed %d calls, want %d", got, iters)
	}
}

type callWaitTestApp struct {
	cfacade.IApplication
}

func (*callWaitTestApp) NodeID() string { return "callwait_test" }

type callWaitTestReceiver struct {
	Base
	count atomic.Int64
}

func (p *callWaitTestReceiver) OnInit() {
	p.Remote().Register("ping", p.ping)
}

func (p *callWaitTestReceiver) ping() int32 {
	p.count.Add(1)
	return ccode.OK
}

type callWaitTestSender struct {
	Base
	iters    int
	failed   atomic.Int64
	lastCode atomic.Int32
	done     chan struct{}
}

func (p *callWaitTestSender) OnInit() {
	p.Remote().Register("run", p.run)
}

func (p *callWaitTestSender) run() {
	defer close(p.done)
	for i := 0; i < p.iters; i++ {
		if code := p.CallWait("callwait_test.receiver", "ping", nil, nil); code != ccode.OK {
			p.failed.Add(1)
			p.lastCode.Store(code)
		}
	}
}
