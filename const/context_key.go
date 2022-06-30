package cherryConst

// context key

type propagateKey struct{}

// PropagateCtxKey is the context key where the content that will be
// propagated through rpc calls is set
var PropagateCtxKey = propagateKey{}

const (
	MessageIdKey = "message_id_key"
)
