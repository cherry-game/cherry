package cherryUtils

import (
	"errors"
	"fmt"
)

var (
	Compress = &compress{}
	Crypto   = &crypto{}
	File     = &file{}
	Reflect  = &reflection{}
	Timer    = &timer{}
	Strings  = &strings{}
	Json     = &json{}
	Net      = &net{}
)

func Error(text string) error {
	return errors.New(text)
}

func ErrorFormat(format string, a ...interface{}) error {
	return errors.New(fmt.Sprintf(format, a))
}
