package cherryHandler

import (
	"context"
	cherryError "github.com/cherry-game/cherry/error"
	ccontext "github.com/cherry-game/cherry/net/context"
	"time"
)

type Executor struct {
	groupIndex int
}

func (p *Executor) SetIndex(index int) {
	p.groupIndex = index
}

func (p *Executor) Index() int {
	return p.groupIndex
}

func (p *Executor) isTimeout(ctx context.Context) error {
	inHandlerTime := ccontext.GetInHandlerTime(ctx)
	duration := time.Duration(time.Now().UnixMilli() - inHandlerTime)
	if duration >= _component.processTimeout {
		return cherryError.Errorf("inHandlerTimeout = %dms", duration)
	}

	buildPacketTime := ccontext.GetBuildPacketTime(ctx)
	duration = time.Duration(time.Now().UnixMilli() - buildPacketTime)
	if duration >= _component.processTimeout {
		return cherryError.Errorf("buildPacketTimeout = %dms", duration)
	}

	return nil
}
