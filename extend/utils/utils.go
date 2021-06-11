package cherryUtils

import (
	"fmt"
)

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
