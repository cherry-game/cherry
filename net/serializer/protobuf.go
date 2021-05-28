package cherrySerializer

import (
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/golang/protobuf/proto"
)

var (
	ErrWrongValueType = cherryUtils.Error("protobuf: convert on wrong type value")
)

// Protobuf implements the serialize.Protobuf interface
type Protobuf struct{}

// NewSerializer returns a new Protobuf.
func NewProtobuf() *Protobuf {
	return &Protobuf{}
}

// Marshal returns the protobuf encoding of v.
func (p *Protobuf) Marshal(v interface{}) ([]byte, error) {
	pb, ok := v.(proto.Message)
	if !ok {
		return nil, ErrWrongValueType
	}
	return proto.Marshal(pb)
}

// Unmarshal parses the protobuf-encoded data and stores the result
// in the value pointed to by v.
func (p *Protobuf) Unmarshal(data []byte, v interface{}) error {
	pb, ok := v.(proto.Message)
	if !ok {
		return ErrWrongValueType
	}
	return proto.Unmarshal(data, pb)
}

// Name returns the name of the serializer.
func (p *Protobuf) Name() string {
	return "protobuf"
}
