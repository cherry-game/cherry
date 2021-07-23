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
		Name() string                                                    // 发现服务名称
		Init(app IApplication, server *grpc.Server, config jsoniter.Any) // 初始化函数
		List() []IMember                                                 // 获取成员列表
		GetType(nodeId string) (nodeType string, err error)              // 根据节点id获取类型
		GetMember(nodeId string) (member IMember, found bool)            // 获取成员
		AddMember(member IMember)                                        // 添加成员
		RemoveMember(nodeId string)                                      // 移除成员
		OnAddMember(listener MemberListener)                             // 添加成员监听函数
		OnRemoveMember(listener MemberListener)                          // 移除成员监听函数
		OnStop()                                                         // 停止当前发现服务
	}

	MemberListener func(member IMember) // MemberListener 成员增、删监听函数
)
