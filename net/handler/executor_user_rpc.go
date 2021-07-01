package cherryHandler

import facade "github.com/cherry-game/cherry/facade"

type (
	UserRPCExecutor struct {
		HandlerFn *facade.HandlerFn
	}
)

func (u *UserRPCExecutor) Invoke() {

}

func (u *UserRPCExecutor) String() string {
	return ""
}
