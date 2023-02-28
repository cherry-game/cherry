package pomeloMessage

import (
	"testing"
)

func TestResponseMessageEncode1(t *testing.T) {
	m := &Message{
		Type:            Response,
		ID:              0,
		Route:           "",
		Data:            []byte(`hello world`),
		routeCompressed: false,
		Error:           true,
	}
	encode, err := Encode(m)
	t.Log(encode, err)

	decode, err := Decode(encode)
	t.Log(decode, err)
}
