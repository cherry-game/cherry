// Package cherryCode defines the framework's numeric status codes and their helpers.
package cherryCode

const (
	OK int32 = 0 // success

	// session / discovery
	SessionUIDNotBind     int32 = 10 // session UID not bound
	DiscoveryNotFoundNode int32 = 11 // target node not found in discovery
	NodeRequestError      int32 = 12 // node request failed

	// RPC transport
	RPCNetError           int32 = 20 // network error
	RPCUnmarshalError     int32 = 21 // unmarshal error
	RPCMarshalError       int32 = 22 // marshal error
	RPCRemoteExecuteError int32 = 23 // remote method execution error

	// Actor execution
	ActorPathIsNil          int32 = 24 // target path is nil
	ActorFuncNameError      int32 = 25 // function name invalid
	ActorConvertPathError   int32 = 26 // path parse error
	ActorMarshalError       int32 = 27 // arg marshal error
	ActorUnmarshalError     int32 = 28 // arg unmarshal error
	ActorInvokeResultIsNil  int32 = 29 // invoke result is nil
	ActorSourceEqualTarget  int32 = 30 // source equals target
	ActorPublishRemoteError int32 = 31 // remote publish error
	ActorChildIDNotFound    int32 = 32 // child Actor ID not found
	ActorCallTimeout        int32 = 33 // call timeout
	ActorIDIsNil            int32 = 34 // Actor ID is nil
	ActorNotFound           int32 = 35 // Actor not found
	ActorInvokeRemoteError  int32 = 36 // remote invoke error
	ActorResponseIsError    int32 = 37 // response is an error
)

// IsOK returns true if code equals OK (0).
func IsOK(code int32) bool {
	return code == OK
}

// IsFail returns true if code is non-zero.
func IsFail(code int32) bool {
	return code != OK
}
