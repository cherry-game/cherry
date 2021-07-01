package proto

type LoginRequest struct {
	Nickname string `json:"nickname"`
}

type LoginResponse struct {
	Code int `json:"code"`
}
type NewUserBroadcast struct {
	Content string `json:"content"`
}

type JoinRoomRequest struct {
	Nickname string `json:"nickname"`
}

type ExistsMembersResponse struct {
	Members string `json:"members"`
}

type SyncMessage struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type UserBalanceResponse struct {
	CurrentBalance int64 `json:"currentBalance"`
}
