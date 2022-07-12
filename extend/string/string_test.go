package cherryString

import (
	"fmt"
	"testing"
)

func TestToInt32(t *testing.T) {
	str1 := ToString(2122003)
	v, ok := ToInt(str1)
	fmt.Println(v, ok)
}

func TestToString(t *testing.T) {
	var valueString = "bbb"
	fmt.Println(ToString(valueString))

	var valueInt = 333
	fmt.Println(ToString(valueInt))

	var valueInt32 int32 = 333
	fmt.Println(ToString(valueInt32))

	var valueInt64 int64 = 333
	fmt.Println(ToString(valueInt64))

	var valueUint uint = 333
	fmt.Println(ToString(valueUint))

	var valueUint32 uint32 = 333
	fmt.Println(ToString(valueUint32))

	var valueUint64 uint64 = 333
	fmt.Println(ToString(valueUint64))

	var valueUint8 uint8 = 10
	fmt.Println(ToString(valueUint8))
}
