package cherryProto

func (x *Member) IsTimeout(nowMills int64) bool {
	return x.LastAt+x.HeartbeatTimeout < nowMills
}

func (x *Member) UpdateSettings(settings map[string]string) {
	if settings == nil {
		return
	}

	for k, v := range settings {
		x.Settings[k] = v
	}
}

func (x *Member) UpdateSetting(key, value string) {
	x.Settings[key] = value
}
