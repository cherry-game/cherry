package cherryMap

import (
	"fmt"
	"testing"
)

func TestMapAll(t *testing.T) {
	sm := NewMap[int32, int32](true)

	sm.Put(1, 1)
	sm.Put(2, 2)
	sm.Put(3, 3)

	for _, k := range sm.Keys() {
		fmt.Println(k)
	}

	for _, v := range sm.Values() {
		fmt.Println(v)
	}

	key1Value, isGet := sm.Get(1)
	fmt.Println(key1Value, isGet)

	sm.Remove(1)

	key1Value, isGet = sm.Get(1)
	fmt.Println(key1Value, isGet)

	size := sm.Size()
	fmt.Println(size)

	isEmpty := sm.Empty()
	fmt.Println(isEmpty)

	sm.Clear()

	sm.Put(4, 4)
	sm.Put(5, 5)
	sm.Put(6, 6)

	for _, k := range sm.Keys() {
		v, _ := sm.Get(k)
		fmt.Printf("k = %d, v = %d \n", k, v)
	}

	fmt.Println(sm.String())

}
