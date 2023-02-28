package test_protobuf

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"testing"
)

func BenchmarkNameGoBool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		boolValue := false
		boolProtoValue := getType(boolValue)
		boolProtoBytes, _ := proto.Marshal(boolProtoValue)
		newBoolProto := &GoBool{}
		proto.Unmarshal(boolProtoBytes, newBoolProto)
	}
}

func BenchmarkGoBool1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		boolValue := false
		boolProtoValue := &GoBool{
			Value: boolValue,
		}
		boolProtoBytes, _ := proto.Marshal(boolProtoValue)
		newBoolProto := &GoBool{}
		proto.Unmarshal(boolProtoBytes, newBoolProto)
	}
}

func TestName(t *testing.T) {
	boolValue := false
	boolProtoValue := getType(boolValue)
	boolProtoBytes, _ := proto.Marshal(boolProtoValue)
	newBoolProto := &GoBool{}
	proto.Unmarshal(boolProtoBytes, newBoolProto)

	stringValue := "string+value"
	stringProtoValue := getType(stringValue)
	stringProtoBytes, _ := proto.Marshal(stringProtoValue)
	newStringProto := &GoString{}
	proto.Unmarshal(stringProtoBytes, newStringProto)

	var int32Value int32 = 1111
	int32ProtoValue := getType(int32Value)
	int32ProtoBytes, _ := proto.Marshal(int32ProtoValue)
	int32Proto := &GoInt32{}
	proto.Unmarshal(int32ProtoBytes, int32Proto)

	fmt.Println(boolValue, stringValue, int32Value)

	fmt.Println(getType(boolValue))
	fmt.Println(getType(stringValue))
	fmt.Println(getType(int32Value))

}

func getType(i interface{}) proto.Message {
	switch t := i.(type) {
	case bool:
		{
			return &GoBool{
				Value: t,
			}
		}
	case string:
		{
			return &GoString{
				Value: t,
			}
		}
	case int32:
		{
			return &GoInt32{
				Value: t,
			}
		}
	case proto.Message:
		{
			return t
		}
	}

	return nil
}
