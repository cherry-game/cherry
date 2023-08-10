package cherrySlice

import (
	"fmt"
	"testing"
)

func TestUnique(t *testing.T) {
	list := Unique[string]("1", "2", "3", "1")
	fmt.Println(list)
}

func TestUniques(t *testing.T) {
	s1 := []string{"1", "2", "3"}
	s2 := []string{"1", "2", "3"}

	list := Uniques[string](s1, s2)
	fmt.Println(list)
}
