package cherryNats

import (
	"testing"

	"github.com/nats-io/nats.go"
)

func TestNatsMsgReleaseDoesNotMutateCallerHeader(t *testing.T) {
	header := nats.Header{
		REQ_ID: {"req-1"},
	}

	msg := GetNatsMsg()
	msg.Header = header
	msg.Subject = "subject"
	msg.Reply = "reply"
	msg.Data = []byte("payload")
	ReleaseNatsMsg(msg)

	if got := header.Get(REQ_ID); got != "req-1" {
		t.Fatalf("expected caller header to stay intact, got %q", got)
	}

	pooled := GetNatsMsg()
	defer ReleaseNatsMsg(pooled)

	if pooled.Header == nil {
		t.Fatal("expected pooled message header to be reinitialized")
	}
	if got := pooled.Header.Get(REQ_ID); got != "" {
		t.Fatalf("expected pooled header to be empty, got %q", got)
	}
	if pooled.Subject != "" {
		t.Fatalf("expected pooled subject to be reset, got %q", pooled.Subject)
	}
	if pooled.Reply != "" {
		t.Fatalf("expected pooled reply to be reset, got %q", pooled.Reply)
	}
	if pooled.Data != nil {
		t.Fatal("expected pooled data to be reset")
	}
}
