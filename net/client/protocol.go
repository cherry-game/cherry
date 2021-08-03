package cherryClient

type loginRequest struct {
	Nickname string `json:"nickname"`
}

type loginResponse struct {
	Code int `json:"code"`
}
type newUserBroadcast struct {
	Content string `json:"content"`
}

type joinRoomRequest struct {
	Nickname string `json:"nickname"`
}

type existsMembersResponse struct {
	Members string `json:"members"`
}

type syncMessage struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type userBalanceResponse struct {
	CurrentBalance int64 `json:"currentBalance"`
}
