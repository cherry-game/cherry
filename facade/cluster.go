package cherryFacade

import (
	cproto "github.com/cherry-game/cherry/net/proto"
	"time"
)

type (
	// IDiscovery 发现服务接口
	IDiscovery interface {
		Load(app IApplication)
		Name() string                                                 // 发现服务名称
		List() []IMember                                              // 获取成员列表
		ListByType(nodeType string, filterNodeId ...string) []IMember // 根据节点类型获取列表
		GetType(nodeId string) (nodeType string, err error)           // 根据节点id获取类型
		GetMember(nodeId string) (member IMember, found bool)         // 获取成员
		AddMember(member IMember)                                     // 添加成员
		RemoveMember(nodeId string)                                   // 移除成员
		OnAddMember(listener MemberListener)                          // 添加成员监听函数
		OnRemoveMember(listener MemberListener)                       // 移除成员监听函数
		Stop()
	}

	IMember interface {
		GetNodeId() string
		GetNodeType() string
		GetAddress() string
		GetSettings() map[string]string
	}

	MemberListener func(member IMember) // MemberListener 成员增、删监听函数
)

type (
	ICluster interface {
		Init()                                                                                               // 初始化
		PublishLocal(nodeId string, packet *cproto.ClusterPacket) error                                      // 发布本地消息
		PublishRemote(nodeId string, packet *cproto.ClusterPacket) error                                     // 发布远程消息
		RequestRemote(nodeId string, packet *cproto.ClusterPacket, timeout ...time.Duration) cproto.Response // 请求远程消息
		Stop()                                                                                               // 停止
	}
)
