package cherryLogger

import "go.uber.org/zap/zapcore"

type Wrapper interface {
	Wrap(core zapcore.Core) zapcore.Core
}

var wrappers []Wrapper

func RegisterWrapper(wrapper Wrapper) {
	wrappers = append(wrappers, wrapper)
}
