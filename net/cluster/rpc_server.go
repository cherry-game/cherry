package cherryCluster

import (
	"github.com/cherry-game/cherry/facade"
)

type RPCServer struct {
}

func (*RPCServer) All() cherryFacade.NodeMap {
	return nil
}

func (*RPCServer) GetType(nodeId string) (nodeType string, err error) {
	return "", nil
}

func (*RPCServer) Get(nodeId string) cherryFacade.INode {
	return nil
}

func (*RPCServer) Sync() {

}

func (*RPCServer) AddListener(listener cherryFacade.INodeListener) {

}
