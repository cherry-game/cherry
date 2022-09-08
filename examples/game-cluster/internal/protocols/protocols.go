package protocols

type LoginRequest struct {
	ServerId string `json:"server_id"`
	Token    string `json:"token"`
}

type LoginResponse struct {
	Uid int64 `json:"uid"`
}

type ActorSelectRequest struct {
	Time int `json:"time"`
}

type ActorSelectResponse struct {
	ActorName string `json:"actor_name"`
}

type ActorCreateRequest struct {
	ActorName string `json:"actor_name"`
	Sex       int    `json:"sex"`
}

type ActorCreateResponse struct {
	ActorInfo *ActorInfo
}

type ActorInfo struct {
	ActorId    int64  `json:"actor_id"`
	ActorName  string `json:"actor_name"`
	ActorSex   int32  `json:"actor_sex"`
	CreateTime int64  `json:"create_time"`
}
