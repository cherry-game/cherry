package cherryProto

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"testing"
)

func TestMarshal(t *testing.T) {
	req1 := &Request{
		Sid:        "11111111",
		Uid:        1,
		FrontendId: "1",
		MsgType:    1,
		MsgId:      2,
		Route:      "11",
		IsError:    true,
		Data:       []byte{},
	}

	bytes, err := proto.Marshal(req1)
	fmt.Println(err)
	fmt.Println(len(bytes))

	req2 := &Request{}
	proto.Unmarshal(bytes, req2)
	fmt.Println(req2)
}
