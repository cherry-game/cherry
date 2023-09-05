package cherryProto

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestMarshal(t *testing.T) {
	req1 := &ClusterPacket{
		SourcePath: "",
		TargetPath: "",
		FuncName:   "",
		ArgBytes:   nil,
		Session:    nil,
	}

	bytes, err := proto.Marshal(req1)
	fmt.Println(err)
	fmt.Println(len(bytes))

	req2 := &ClusterPacket{}
	proto.Unmarshal(bytes, req2)
	fmt.Println(req2)
}
