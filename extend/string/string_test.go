package cherryString

import (
	"fmt"
	"testing"
)

func TestToInt32(t *testing.T) {
	str1 := Int32ToString(2122003)
	v, ok := ToInt(str1)
	fmt.Println(v, ok)
}
