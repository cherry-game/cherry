package cherryActor

import (
	cerror "github.com/cherry-game/cherry/error"
)

var (
	ErrForbiddenToCallSelf       = cerror.Errorf("SendActorID cannot be equal to TargetActorID")
	ErrForbiddenCreateChildActor = cerror.Errorf("Forbidden create child actor")
	ErrActorIDIsNil              = cerror.Error("actorID is nil.")
	ErrNotChildActorID           = cerror.Error("Not child actorID")
	ErrActorPathTell             = cerror.Error("Target Path tell error.")
)

const (
	LocalName  = "local"
	RemoteName = "remote"
)
