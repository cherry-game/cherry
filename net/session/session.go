package cherrySession

//import (
//	"fmt"
//	cnuid "github.com/cherry-game/cherry/extend/nuid"
//	cstring "github.com/cherry-game/cherry/extend/string"
//	cfacade "github.com/cherry-game/cherry/facade"
//	clog "github.com/cherry-game/cherry/logger"
//	"go.uber.org/zap/zapcore"
//)
//
//type Session map[string]string

//
//func NewSession(frontendId cfacade.FrontendID, gateActorId string) Session {
//	sid := cnuid.Next()
//	session := Session{}
//	session[KeySID] = sid
//	session[KeyFrontendID] = frontendId
//	session[KeyAgentPath] = fmt.Sprintf("%s.%s.%s", frontendId, gateActorId, sid)
//	session[KeyIP] = ""
//	return session
//}
//
//func (m Session) SID() cfacade.SID {
//	return m.GetString(KeySID)
//}
//
//func (m Session) UID() cfacade.UID {
//	return m.GetInt64(KeyUID)
//}
//
//func (m Session) FrontendID() cfacade.FrontendID {
//	return m.GetString(KeyFrontendID)
//}
//
//func (m Session) AgentPath() string {
//	return m.GetString(KeyAgentPath)
//}
//
//func (m Session) IP() string {
//	return m.GetString(KeyIP)
//}
//
//func (m Session) MessageID() uint {
//	return m.GetUint(KeyMessageID)
//}
//
//func (m Session) PacketTime() int64 {
//	return m.GetInt64(KeyPacketTime)
//}
//
//func (m Session) IsBind() bool {
//	return m.UID() > 0
//}
//
//func (m Session) Remove(key string) {
//	delete(m, key)
//}
//
//func (m Session) ImportAll(data map[string]string) {
//	for k, v := range data {
//		m.Set(k, v)
//	}
//}
//
//func (m Session) Set(key string, value string) {
//	if key == "" || value == "" {
//		return
//	}
//
//	m[key] = value
//}
//
//func (m Session) Contains(key string) bool {
//	_, found := m[key]
//	return found
//}
//
//func (m Session) Restore(data map[string]string) {
//	m.Clear()
//
//	for k, v := range data {
//		m.Set(k, v)
//	}
//}
//
//// Clear releases all settings related to current sc
//func (m Session) Clear() {
//	for k := range m {
//		delete(m, k)
//	}
//}
//
//func (m Session) AddData(key string, value interface{}) {
//	m[key] = cstring.ToString(value)
//}
//
//func (m Session) GetUint(key string) uint {
//	v, ok := m[key]
//	if !ok {
//		return 0
//	}
//
//	value, ok := cstring.ToUint(v)
//	if !ok {
//		return 0
//	}
//	return value
//}
//
//func (m Session) GetInt(key string) int {
//	v, ok := m[key]
//	if !ok {
//		return 0
//	}
//
//	value, ok := cstring.ToInt(v)
//	if !ok {
//		return 0
//	}
//	return value
//}
//
//// GetInt32 returns the value associated with the key as a int32.
//func (m Session) GetInt32(key string) int32 {
//	v, ok := m[key]
//	if !ok {
//		return 0
//	}
//
//	value, ok := cstring.ToInt32(v)
//	if !ok {
//		return 0
//	}
//	return value
//}
//
//// GetInt64 returns the value associated with the key as a int64.
//func (m Session) GetInt64(key string) int64 {
//	v, ok := m[key]
//	if !ok {
//		return 0
//	}
//
//	value, ok := cstring.ToInt64(v)
//	if !ok {
//		return 0
//	}
//	return value
//}
//
//// GetString returns the value associated with the key as a string.
//func (m Session) GetString(key string) string {
//	v, ok := m[key]
//	if !ok {
//		return ""
//	}
//
//	return v
//}
//
//func (m Session) Copy() Session {
//	session := Session{}
//	for k, v := range m {
//		session.Set(k, v)
//	}
//	return session
//}
//
//func (m Session) logPrefix() string {
//	return fmt.Sprintf("[uid = %d] ", m.UID())
//}
//
//func (m Session) Debug(args ...interface{}) {
//	clog.DefaultLogger.Debug(m.logPrefix(), fmt.Sprint(args...))
//}
//
//func (m Session) Debugf(template string, args ...interface{}) {
//	clog.DefaultLogger.Debug(m.logPrefix(), fmt.Sprintf(template, args...))
//}
//
//func (m Session) Info(args ...interface{}) {
//	clog.DefaultLogger.Info(m.logPrefix(), fmt.Sprint(args...))
//}
//
//func (m Session) Infof(template string, args ...interface{}) {
//	clog.DefaultLogger.Info(m.logPrefix(), fmt.Sprintf(template, args...))
//}
//
//func (m Session) Warn(args ...interface{}) {
//	clog.DefaultLogger.Warn(m.logPrefix(), fmt.Sprint(args...))
//}
//
//func (m Session) Warnf(template string, args ...interface{}) {
//	clog.DefaultLogger.Warn(m.logPrefix(), fmt.Sprintf(template, args...))
//}
//
//func (m Session) Error(args ...interface{}) {
//	clog.DefaultLogger.Error(m.logPrefix(), fmt.Sprint(args...))
//}
//
//func (m Session) Errorf(template string, args ...interface{}) {
//	clog.DefaultLogger.Error(m.logPrefix(), fmt.Sprintf(template, args...))
//}
//
//func (m Session) LogEnable(level zapcore.Level) bool {
//	return clog.Enable(level)
//}
