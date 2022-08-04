package cherryUtils

import (
	"fmt"
	"reflect"
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

// IsPtr is ptr type
func IsPtr(value interface{}) bool {
	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Ptr {
		return true
	}

	return false
}
