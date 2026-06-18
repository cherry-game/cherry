// Package cherryError provides lightweight error constructors and sentinel errors
// used throughout the Cherry framework.
//
// Constructors:
//   - Error / Errorf — create a new error
//   - Wrap / Wrapf — wrap an existing error with context
//
// Sentinel errors are grouped by subsystem:
//   - route, packet, message, protobuf, cluster, discovery, actor, func
package cherryError

import (
	"errors"
	"fmt"
)

// Error creates a new error from a static string.
func Error(text string) error {
	return errors.New(text)
}

// Errorf creates a new error from a format string and args.
func Errorf(format string, a ...interface{}) error {
	return Error(fmt.Sprintf(format, a...))
}

// Wrap wraps an existing error with additional context text.
func Wrap(err error, text string) error {
	return Errorf("err:%v, text:%s", err, text)
}

// Wrapf wraps an existing error with formatted context text.
func Wrapf(err error, format string, a ...interface{}) error {
	text := fmt.Sprintf(format, a...)
	return Wrap(err, text)
}

// --- sentinel errors ---

// route
var (
	RouteFieldCantEmpty = Error("Route field cannot be empty")
	RouteInvalid        = Error("Invalid route")
)

// packet
var (
	PacketWrongType              = Error("Wrong packet type")
	PacketSizeExceed             = Error("Codec: packet size exceeded")
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

// protobuf
var (
	ProtobufWrongValueType = Error("Conversion on wrong value type")
)

// cluster
var (
	ClusterClientIsStop           = Error("Cluster client is stopped")
	ClusterRequestTimeout         = Error("Cluster request timeout")
	ClusterPacketMarshalFail      = Error("Cluster packet marshal failed")
	ClusterPacketUnmarshalFail    = Error("Cluster packet unmarshal failed")
	ClusterPublishFail            = Error("Cluster publish failed")
	ClusterRequestFail            = Error("Cluster request failed")
	ClusterNodeTypeIsNil          = Error("Cluster node type is nil")
	ClusterNodeTypeMemberNotFound = Error("Cluster node type member not found")
)

// discovery
var (
	DiscoveryNotFoundNode = Error("Discovery node not found")
)

// actor
var (
	ActorPathError = Error("Actor path is invalid")
)

// func
var (
	FuncIsNil     = Error("Func is nil")
	FuncTypeError = Error("Func type error")
)
