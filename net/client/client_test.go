package cherryClient

import (
	"testing"
	"time"
)

func TestClient(t *testing.T) {

	c := New(100 * time.Millisecond)

	c.ConnectTo("127.0.0.1:34590")

	defer c.Disconnect()

}
