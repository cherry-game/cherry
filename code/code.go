package cherryCode

const (
	OK                    int32 = 0  // is ok
	SessionUIDNotBind     int32 = 10 // session uid not bind
	DiscoveryNotFoundNode int32 = 11 // discovery not fond node id
	NodeRequestError      int32 = 12 // node request error
	RPCNetError           int32 = 20 // rpc net error
	RPCUnmarshalError     int32 = 21 // rpc data unmarshal error
	RPCMarshalError       int32 = 22 // rpc data marshal error
	RPCRemoteExecuteError int32 = 23 // rpc remote method executor error

	ActorTargetPathIsNil    int32 = 24 // actor target path is nil
	ActorFuncNameError      int32 = 25 // actor function name is error
	ActorConvertPathError   int32 = 26 // convert to path error
	ActorMarshalError       int32 = 27 // marshal arg error
	ActorUnmarshalError     int32 = 28 // unmarshal arg error
	ActorCallFail           int32 = 29 // actor call fail
	ActorSourceEqualTarget  int32 = 30 // source equal target
	ActorPublishRemoteError int32 = 31 // actor publish remote error

)

func IsOK(code int32) bool {
	return code == OK
}

func IsFail(code int32) bool {
	return code != OK
}
