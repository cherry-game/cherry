package cherryFacade

import (
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc"
)

type (
	IMember interface {
		GetNodeId() string
		GetNodeType() string
		GetAddress() string
		GetSettings() map[string]string
	}

	// IDiscovery 节点发现接口
	IDiscovery interface {
		Name() string
		Init(app IApplication, rpcServer *grpc.Server, discoveryConfig jsoniter.Any)
		List() []IMember
		GetType(nodeId string) (nodeType string, err error)
		GetMember(nodeId string) (member IMember, found bool)
		AddMember(member IMember)
		RemoveMember(nodeId string)
		OnAddMember(listener MemberListener)
		OnRemoveMember(listener MemberListener)
		OnStop()
	}

	// MemberListener 成员增、删监听函数
	MemberListener func(member IMember)
)
