package cherryHandler

import facade "github.com/cherry-game/cherry/facade"

type (
	SysRPCExecutor struct {
		HandlerFn *facade.HandlerFn
	}
)

func (s *SysRPCExecutor) Invoke() {

}

func (s *SysRPCExecutor) String() string {
	return ""
}
