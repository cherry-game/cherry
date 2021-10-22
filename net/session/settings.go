package cherrySession

import (
	cherryString "github.com/cherry-game/cherry/extend/string"
	"sync"
)

type settings struct {
	sync.RWMutex
	data map[string]string // setting data
}

func (s *settings) Data() map[string]string {
	s.RLock()
	defer s.RUnlock()
	return s.data
}

func (s *settings) Remove(key string) {
	s.Lock()
	defer s.Unlock()

	delete(s.data, key)
}

func (s *settings) ImportAll(settings map[string]string) {
	for k, v := range settings {
		s.Set(k, v)
	}
}

func (s *settings) Set(key string, value string) {
	if key == "" || value == "" {
		return
	}

	s.Lock()
	defer s.Unlock()
	s.data[key] = value
}

func (s *settings) Contains(key string) bool {
	s.RLock()
	defer s.RUnlock()

	_, exist := s.data[key]
	return exist
}

func (s *settings) Restore(data map[string]string) {
	s.Lock()
	defer s.Unlock()

	s.data = data
}

// Clear releases all settings related to current sc
func (s *settings) Clear() {
	s.Lock()
	defer s.Unlock()

	s.data = map[string]string{}
}

func (s *settings) GetInt(key string) int {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := cherryString.ToInt(v)
	if !ok {
		return 0
	}
	return value
}

// GetInt32 returns the value associated with the key as a int32.
func (s *settings) GetInt32(key string) int32 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := cherryString.ToInt32(v)
	if !ok {
		return 0
	}
	return value
}

// GetInt64 returns the value associated with the key as a int64.
func (s *settings) GetInt64(key string) int64 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := cherryString.ToInt64(v)
	if !ok {
		return 0
	}
	return value
}

// GetString returns the value associated with the key as a string.
func (s *settings) GetString(key string) string {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return ""
	}

	return v
}
