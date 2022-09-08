package main

// 登录请求
type loginRequest struct {
	Nickname string `json:"nickname"`
}

// 登录响应
type loginResponse struct {
	Code int `json:"code"`
}

// 新用户广播
type newUserBroadcast struct {
	Content string `json:"content"`
}

// 加入房间请求
type joinRoomRequest struct {
	Nickname string `json:"nickname"`
}

// 存在的角色响应
type existsMembersResponse struct {
	Members string `json:"members"`
}

// 同步聊天消息
type syncMessage struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// 用户聊天次数响应
type userBalanceResponse struct {
	CurrentBalance int64 `json:"currentBalance"`
}
