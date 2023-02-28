package pomeloClient

import (
	"fmt"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	client := New(
		WithRequestTimeout(1 * time.Second),
	)
	client.TagName = "dog egg"
	client.ConnectToWS("172.16.124.137:21000", "")

	defer client.Disconnect()

	time.Sleep(1 * time.Hour)

}

func BenchmarkClient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		client := New(
			WithRequestTimeout(1 * time.Second),
		)
		client.TagName = fmt.Sprintf("c-%d", i)
		client.ConnectToWS("172.16.124.137:21000", "")

		client.Disconnect()
	}
}
