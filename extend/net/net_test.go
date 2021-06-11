package cherryNet

import (
	"fmt"
	"testing"
)

func TestLocalIPV4(t *testing.T) {
	ip := LocalIPV4()
	fmt.Println(ip)
}
