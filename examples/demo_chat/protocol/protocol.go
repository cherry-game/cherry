package protocol

import cherryTime "github.com/cherry-game/cherry/extend/time"

// LoginRequest 登录请求
type LoginRequest struct {
	Nickname string `json:"nickname"`
	UID      int64  `json:"uid"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Code int `json:"code"`
}

// NewUserBroadcast 新用户广播
type NewUserBroadcast struct {
	Content string `json:"content"`
}

// ExistsMembersResponse 存在的角色响应
type ExistsMembersResponse struct {
	Members string `json:"members"`
}

// SyncMessage 同步聊天消息
type SyncMessage struct {
	Name       string `json:"name"`
	Content    string `json:"content"`
	PacketTime int64  `json:"packetTime"`
}

// UserBalanceResponse 用户聊天次数响应
type UserBalanceResponse struct {
	CurrentBalance int64 `json:"currentBalance"`
}

type WriteResponse struct {
	Result bool
}

type Int64 struct {
	Value int64 `json:"value"`
}

// PacketSpendTime 从打包到当前时间，花费了多少时间(纳秒)
func (p *SyncMessage) PacketSpendTime() int64 {
	return cherryTime.Now().UnixNano() - p.PacketTime
}
