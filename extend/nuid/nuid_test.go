package cherryNUID

import (
	"fmt"
	"testing"
)

func TestNUID(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(id.Next())
	}
}

func BenchmarkNUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		id.Next()
	}
}
