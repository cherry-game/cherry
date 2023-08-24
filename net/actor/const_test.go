package cherryActor

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
)

type testActor struct {
	Base
}

func TestActorSystem(t *testing.T) {
	actorSystem := NewSystem()

	ta := &testActor{}
	actorSystem.CreateActor("aaa", ta)

	fmt.Println(ta.Base.path)

}

func Test1111(t *testing.T) {
	a := [...]int{1, 2, 3}
	square(&a)

	fmt.Println(a)
}

func square(arr *[3]int) {
	for i, num := range *arr {
		(*arr)[i] = num * num
	}
}

func TestActorIDValidate1(t *testing.T) {
	s := "/aaa/bbb"
	str := "-"

	index := strings.Index(s, str)
	fmt.Println(index)
}

func TestActorIDValidate(t *testing.T) {
	idList := []string{
		"",
		" ",
		"@",
		"/",
		"a/1",
		"a",
		"A",
		"1",
		"aaaa1111",
		"1111aaaaa",
		"a.b.c",
	}

	checkActorID := func(id string) bool {
		return len(id) > 0
	}

	for _, s := range idList {
		fmt.Println(s, "->", checkActorID(s))
	}
}

func TestUint32(t *testing.T) {
	var id uint32 = 4294967290

	for i := 0; i < 20; i++ {
		atomic.AddUint32(&id, 1)
		fmt.Println(id)
	}
}
