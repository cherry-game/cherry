package cherryCluster

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/facade"
)

type RPCServerComponent struct {
	cherryFacade.Component
}

func (p *RPCServerComponent) Name() string {
	return cherryConst.RPCServerComponent
}

func (p *RPCServerComponent) Init() {

}

func (p *RPCServerComponent) OnAfterInit() {

}

func (p *RPCServerComponent) OnStop() {

}

func (*RPCServerComponent) All() cherryFacade.NodeMap {
	return nil
}

func (*RPCServerComponent) GetType(nodeId string) (nodeType string, err error) {
	return "", nil
}

func (*RPCServerComponent) Get(nodeId string) cherryFacade.INode {
	return nil
}

func (*RPCServerComponent) Sync() {

}

func (*RPCServerComponent) AddListener(listener cherryFacade.INodeListener) {

}
