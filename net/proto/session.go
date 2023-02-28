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

func (m *Session) IsBind() bool {
	return m.Uid > 0
}

func (m *Session) ActorPath() string {
	return m.AgentPath + cconst.DOT + m.Sid
}

func (m *Session) Add(key string, value interface{}) {
	m.Data[key] = cstring.ToString(value)
}

func (m *Session) Remove(key string) {
	delete(m.Data, key)
}

func (m *Session) Set(key string, value string) {
	if key == "" || value == "" {
		return
	}

	m.Data[key] = value
}

func (m *Session) ImportAll(data map[string]string) {
	for k, v := range data {
		m.Set(k, v)
	}
}

func (m *Session) Contains(key string) bool {
	_, found := m.Data[key]
	return found
}

func (m *Session) Restore(data map[string]string) {
	m.Clear()

	for k, v := range data {
		m.Set(k, v)
	}
}

// Clear releases all settings related to current sc
func (m *Session) Clear() {
	for k := range m.Data {
		delete(m.Data, k)
	}
}

func (m *Session) GetUint(key string) uint {
	v, ok := m.Data[key]
	if !ok {
		return 0
	}

	value, ok := cstring.ToUint(v)
	if !ok {
		return 0
	}
	return value
}

func (m *Session) GetInt(key string) int {
	v, ok := m.Data[key]
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
func (m *Session) GetInt32(key string) int32 {
	v, ok := m.Data[key]
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
func (m *Session) GetInt64(key string) int64 {
	v, ok := m.Data[key]
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
func (m *Session) GetString(key string) string {
	v, ok := m.Data[key]
	if !ok {
		return ""
	}

	return v
}

func (m *Session) Copy() Session {
	session := NewSession(m.Sid, m.AgentPath)
	session.Uid = m.Uid
	session.Ip = m.Ip
	for k, v := range m.Data {
		session.Set(k, v)
	}
	return session
}
