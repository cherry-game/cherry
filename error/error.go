package cherryError

import (
	"errors"
	"fmt"
)

func Error(text string) error {
	return errors.New(text)
}

func Errorf(format string, a ...interface{}) error {
	return Error(fmt.Sprintf(format, a...))
}

func Wrap(err error, text string) error {
	return Errorf("err:%v, text:%s", err, text)
}

func Wrapf(err error, format string, a ...interface{}) error {
	text := fmt.Sprintf(format, a...)
	return Wrap(err, text)
}

// route
var (
	RouteFieldCantEmpty = Error("Route field can not be empty")
	RouteInvalid        = Error("Invalid route")
)

// packet
var (
	PacketWrongType              = Error("Wrong packet type")
	PacketSizeExceed             = Error("Codec: packet size exceed")
	PacketConnectClosed          = Error("Client connection closed")
	PacketInvalidHeader          = Error("Invalid header")
	PacketMsgSmallerThanExpected = Error("Received less data than expected, EOF?")
)

// message
var (
	MessageWrongType     = Error("Wrong message type")
	MessageInvalid       = Error("Invalid message")
	MessageRouteNotFound = Error("Route info not found in dictionary")
)

var (
	ProtobufWrongValueType = Error("Convert on wrong type value")
)

// cluster
var (
	ClusterClientIsStop           = Error("Cluster client is stop")
	ClusterRequestTimeout         = Error("Cluster Request timeout")
	ClusterPacketMarshalFail      = Error("Cluster packet marshal fail")
	ClusterPacketUnmarshalFail    = Error("Cluster packet unmarshal fail")
	ClusterPublishFail            = Error("Cluster publish fail")
	ClsuterRequestFail            = Error("Cluster request fail")
	ClusterNodeTypeIsNil          = Error("Cluster node type is nil")
	ClusterNodeTypeMemberNotFound = Error("Cluster node type member not found")
)

var (
	DiscoveryNotFoundNode = Error("Discovery not found node")
)

var (
	ActorPathError = Error("Actor path is error.")
)

var (
	FuncIsNil     = Error("Func is nil")
	FuncTypeError = Error("Func type error")
)
