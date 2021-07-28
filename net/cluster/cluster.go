package cherryCluster

import (
	cherryFacade "github.com/cherry-game/cherry/facade"
)

// IBindStorage 绑定存储，用于存储UID对应的前端节点id
type IBindStorage interface {
	GetFrontendID(uid cherryFacade.UID, nodeType string) (string, error)
	Binding(uid cherryFacade.UID) error
}
