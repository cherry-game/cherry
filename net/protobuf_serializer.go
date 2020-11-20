package cherryNet

import (
	"github.com/cherry-game/cherry/utils"
	"github.com/golang/protobuf/proto"
)

var (
	ErrWrongValueType = cherryUtils.Error("protobuf: convert on wrong type value")
)

// ProtobufSerializer implements the serialize.ProtobufSerializer interface
type ProtobufSerializer struct{}

// NewSerializer returns a new ProtobufSerializer.
func NewProtobuf() *ProtobufSerializer {
	return &ProtobufSerializer{}
}

// Marshal returns the protobuf encoding of v.
func (p *ProtobufSerializer) Marshal(v interface{}) ([]byte, error) {
	pb, ok := v.(proto.Message)
	if !ok {
		return nil, ErrWrongValueType
	}
	return proto.Marshal(pb)
}

// Unmarshal parses the protobuf-encoded data and stores the result
// in the value pointed to by v.
func (p *ProtobufSerializer) Unmarshal(data []byte, v interface{}) error {
	pb, ok := v.(proto.Message)
	if !ok {
		return ErrWrongValueType
	}
	return proto.Unmarshal(data, pb)
}

// Name returns the name of the serializer.
func (p *ProtobufSerializer) Name() string {
	return "protobuf"
}
