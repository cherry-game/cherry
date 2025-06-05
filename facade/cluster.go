package cherryFacade

import (
	"time"

	cproto "github.com/cherry-game/cherry/net/proto"
)

type (
	// IDiscovery 发现服务接口
	IDiscovery interface {
		Load(app IApplication)
		Name() string                                                 // 发现服务名称
		Map() map[string]IMember                                      // 获取成员列表
		ListByType(nodeType string, filterNodeID ...string) []IMember // 根据节点类型获取列表
		Random(nodeType string) (IMember, bool)                       // 根据节点类型随机一个
		GetType(nodeID string) (nodeType string, err error)           // 根据节点id获取类型
		GetMember(nodeID string) (member IMember, found bool)         // 获取成员
		AddMember(member IMember)                                     // 添加成员
		RemoveMember(nodeID string)                                   // 移除成员
		OnAddMember(listener MemberListener)                          // 添加成员监听函数
		OnRemoveMember(listener MemberListener)                       // 移除成员监听函数
		Stop()
	}

	IMember interface {
		GetNodeID() string
		GetNodeType() string
		GetAddress() string
		GetSettings() map[string]string
	}

	MemberListener func(member IMember) // MemberListener 成员增、删监听函数
)

type (
	ICluster interface {
		Init()                                                                                               // 初始化
		PublishLocal(nodeID string, packet *cproto.ClusterPacket) error                                      // 发布本地消息
		PublishRemote(nodeID string, packet *cproto.ClusterPacket) error                                     // 发布远程消息
		RequestRemote(nodeID string, packet *cproto.ClusterPacket, timeout ...time.Duration) cproto.Response // 请求远程消息
		Stop()                                                                                               // 停止
	}
)
