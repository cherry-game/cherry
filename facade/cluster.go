package cherryFacade

import (
	cherryProto "github.com/cherry-game/cherry/net/proto"
	"time"
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
		Name() string                                                 // 发现服务名称
		Init(app IApplication)                                        // 初始化函数
		List() []IMember                                              // 获取成员列表
		ListByType(nodeType string, filterNodeId ...string) []IMember // 根据节点类型获取列表
		GetType(nodeId string) (nodeType string, err error)           // 根据节点id获取类型
		GetMember(nodeId string) (member IMember, found bool)         // 获取成员
		AddMember(member IMember)                                     // 添加成员
		RemoveMember(nodeId string)                                   // 移除成员
		OnAddMember(listener MemberListener)                          // 添加成员监听函数
		OnRemoveMember(listener MemberListener)                       // 移除成员监听函数
		OnStop()                                                      // 停止当前发现服务
	}

	MemberListener func(member IMember) // MemberListener 成员增、删监听函数
)

type (
	rpc interface {
		Init(app IApplication)
		OnStop()
	}

	RPCClient interface {
		rpc
		CallLocal(nodeId string, packet *cherryProto.LocalPacket) error
		CallRemote(nodeId string, route string, val interface{}, timeout time.Duration, rsp *cherryProto.Response)
		CallRemoteAsync(nodeId string, route string, val interface{})
		SendKick(nodeId string, uid UID, reason interface{})
		SendPush(nodeId string, route string, uid UID, val interface{})
	}

	RPCServer interface {
		rpc
	}
)
