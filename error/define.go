package cherryError

// cluster
var (
	ClusterBrokenPipe         = Error("broken low-level pipe")
	ClusterBufferExceed       = Error("session send buffer exceed")
	ClusterErrSessionOnNotify = Error("error session on notify")
	ClusterClosedSession      = Error("session is closed")
	ClusterRPCHandleNotFound  = Error("rpc handler is nil")
)

// session
var (
	SessionIllegalUID = Error("illegal uid")
)

// route
var (
	RouteFieldCantEmpty = Error("route field can not be empty")
	RouteInvalid        = Error("invalid route")
)

// packet
var (
	PacketWrongType              = Error("wrong packet type")
	PacketSizeExceed             = Error("codec: packet size exceed")
	PacketConnectClosed          = Error("client connection closed")
	PacketInvalidHeader          = Error("invalid header")
	PacketMsgSmallerThanExpected = Error("received less data than expected, EOF?")
)

// message
var (
	MessageWrongType     = Error("wrong message type")
	MessageInvalid       = Error("invalid message")
	MessageRouteNotFound = Error("route info not found in dictionary")
)

// serializer protobuf
var (
	ProtobufWrongValueType = Error("convert on wrong type value")
)
