package chat

type LoginRequest struct {
	Nickname string `json:"nickname"`
}

type LoginResponse struct {
	Code int `json:"code"`
}
type NewUserBroadcast struct {
	Content string `json:"content"`
}

type NewUserRequest struct {
	Nickname string `json:"nickname"`
	GateUid  int64  `json:"gateUid"`
}

type JoinRoomRequest struct {
	Nickname  string `json:"nickname"`
	GateUid   int64  `json:"gateUid"`
	MasterUid int64  `json:"masterUid"`
}

type MasterStats struct {
	Uid int64 `json:"uid"`
}

type ExistsMembersResponse struct {
	Members string `json:"members"`
}

type SyncMessage struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}
