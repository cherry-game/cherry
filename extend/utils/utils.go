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
	Slice    = &slice{}
)

func Error(text string) error {
	return errors.New(text)
}

func Errorf(format string, a ...interface{}) error {
	return errors.New(fmt.Sprintf(format, a))
}

func Try(tryFn func(), catchFn func(errString string)) bool {
	var hasException = true
	func() {
		defer catchError(catchFn)
		tryFn()
		hasException = false
	}()
	return hasException
}

func catchError(catch func(errString string)) {
	if r := recover(); r != nil {
		catch(fmt.Sprint(r))
	}
}
