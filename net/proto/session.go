package cherryProto

import (
	cconst "github.com/cherry-game/cherry/const"
	cstring "github.com/cherry-game/cherry/extend/string"
)

func (x *Session) IsBind() bool {
	return x.Uid > 0
}

func (x *Session) ActorPath() string {
	return x.AgentPath + cconst.DOT + x.Sid
}

func (x *Session) Add(key string, value interface{}) {
	x.ensureData()

	x.Data[key] = cstring.ToString(value)
}

func (x *Session) Remove(key string) {
	delete(x.Data, key)
}

func (x *Session) Set(key string, value string) {
	if key == "" || value == "" {
		return
	}

	x.ensureData()

	x.Data[key] = value
}

// ensureData lazily allocates the data map before a write.
//
// A session that crossed the cluster arrives with a nil Data map whenever the
// sender had nothing to send, because proto3 does not serialize an empty map.
// Remote handlers do write session data, so allocate on demand rather than
// panicking with "assignment to entry in nil map".
func (x *Session) ensureData() {
	if x.Data == nil {
		x.Data = map[string]string{}
	}
}

// SetMID sets the message id of the message currently being routed.
//
// mid is per message, not session state: only write it on the per-message
// session copy built for a single incoming packet. Writing it on a session
// shared with another goroutine is a data race.
func (x *Session) SetMID(mid uint32) {
	x.Mid = mid
}

// GetMID returns the message id of the message currently being routed.
func (x *Session) GetMID() uint32 {
	return x.Mid
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

func (x *Session) Equal(key, value string) bool {
	dataValue, found := x.Data[key]
	if !found {
		return false
	}

	return dataValue == value
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
