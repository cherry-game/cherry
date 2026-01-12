package cherryProto

func (x *Member) IsTimeout(nowMills int64) bool {
	return x.LastAt+x.HeartbeatTimeout < nowMills
}
