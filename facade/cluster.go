package cherryFacade

import jsoniter "github.com/json-iterator/go"

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
		Init(app IApplication, discoveryConfig jsoniter.Any)
		List() []IMember
		GetType(nodeId string) (nodeType string, err error)
		Get(nodeId string) (member IMember, found bool)
		OnAddMember(listener MemberListener)
		OnRemoveMember(listener MemberListener)
	}

	// MemberListener 成员增、删监听函数
	MemberListener func(member IMember)
)
