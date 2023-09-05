package cherryProto

import (
	cconst "github.com/cherry-game/cherry/const"
	cstring "github.com/cherry-game/cherry/extend/string"
)

func NewSession(sid, agentPath string) Session {
	session := Session{
		Sid:       sid,
		AgentPath: agentPath,
		Data:      map[string]string{},
	}
	return session
}

func (x *Session) IsBind() bool {
	return x.Uid > 0
}

func (x *Session) ActorPath() string {
	return x.AgentPath + cconst.DOT + x.Sid
}

func (x *Session) Add(key string, value interface{}) {
	x.Data[key] = cstring.ToString(value)
}

func (x *Session) Remove(key string) {
	delete(x.Data, key)
}

func (x *Session) Set(key string, value string) {
	if key == "" || value == "" {
		return
	}

	x.Data[key] = value
}

func (x *Session) ImportAll(data map[string]string) {
	for k, v := range data {
		x.Set(k, v)
	}
}

func (x *Session) Contains(key string) bool {
	_, found := x.Data[key]
	return found
}

func (x *Session) Restore(data map[string]string) {
	x.Clear()

	for k, v := range data {
		x.Set(k, v)
	}
}

// Clear releases all settings related to current sc
func (x *Session) Clear() {
	for k := range x.Data {
		delete(x.Data, k)
	}
}

func (x *Session) GetUint(key string) uint {
	v, ok := x.Data[key]
	if !ok {
		return 0
	}

	value, ok := cstring.ToUint(v)
	if !ok {
		return 0
	}
	return value
}

func (x *Session) GetInt(key string) int {
	v, ok := x.Data[key]
	if !ok {
		return 0
	}

	value, ok := cstring.ToInt(v)
	if !ok {
		return 0
	}
	return value
}

// GetInt32 returns the value associated with the key as a int32.
func (x *Session) GetInt32(key string) int32 {
	v, ok := x.Data[key]
	if !ok {
		return 0
	}

	value, ok := cstring.ToInt32(v)
	if !ok {
		return 0
	}
	return value
}

// GetInt64 returns the value associated with the key as a int64.
func (x *Session) GetInt64(key string) int64 {
	v, ok := x.Data[key]
	if !ok {
		return 0
	}

	value, ok := cstring.ToInt64(v)
	if !ok {
		return 0
	}
	return value
}

// GetString returns the value associated with the key as a string.
func (x *Session) GetString(key string) string {
	v, ok := x.Data[key]
	if !ok {
		return ""
	}

	return v
}

func (x *Session) Copy() Session {
	session := NewSession(x.Sid, x.AgentPath)
	session.Uid = x.Uid
	session.Ip = x.Ip
	for k, v := range x.Data {
		session.Set(k, v)
	}
	return session
}
