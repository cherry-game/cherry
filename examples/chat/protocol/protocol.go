package protocol

// LoginRequest 登录请求
type LoginRequest struct {
	Nickname string `json:"nickname"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Code int `json:"code"`
}

// NewUserBroadcast 新用户广播
type NewUserBroadcast struct {
	Content string `json:"content"`
}

// JoinRoomRequest 加入房间请求
type JoinRoomRequest struct {
	Nickname string `json:"nickname"`
}

// ExistsMembersResponse 存在的角色响应
type ExistsMembersResponse struct {
	Members string `json:"members"`
}

// SyncMessage 同步聊天消息
type SyncMessage struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// UserBalanceResponse 用户聊天次数响应
type UserBalanceResponse struct {
	CurrentBalance int64 `json:"currentBalance"`
}

type WriteResponses struct {
	Result bool `json:"result"`
}
