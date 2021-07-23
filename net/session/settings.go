package cherrySession

import "sync"

type settings struct {
	sync.RWMutex
	data map[string]interface{} // setting data
}

func (s *settings) Data() map[string]interface{} {
	s.RLock()
	defer s.RUnlock()

	return s.data
}

func (s *settings) Remove(key string) {
	s.Lock()
	defer s.Unlock()

	delete(s.data, key)
}

func (s *settings) Set(key string, value interface{}) {
	if key == "" || value == nil {
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

func (s *settings) Restore(data map[string]interface{}) {
	s.Lock()
	defer s.Unlock()

	s.data = data
}

// Clear releases all settings related to current sc
func (s *settings) Clear() {
	s.Lock()
	defer s.Unlock()

	s.data = map[string]interface{}{}
}

func (s *settings) GetInt(key string) int {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(int)
	if !ok {
		return 0
	}
	return value
}

// GetInt8 returns the value associated with the key as a int8.
func (s *settings) GetInt8(key string) int8 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(int8)
	if !ok {
		return 0
	}
	return value
}

// GetInt16 returns the value associated with the key as a int16.
func (s *settings) GetInt16(key string) int16 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(int16)
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

	value, ok := v.(int32)
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

	value, ok := v.(int64)
	if !ok {
		return 0
	}
	return value
}

// GetUint returns the value associated with the key as a uint.
func (s *settings) GetUint(key string) uint {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(uint)
	if !ok {
		return 0
	}
	return value
}

// GetUint8 returns the value associated with the key as a uint8.
func (s *settings) GetUint8(key string) uint8 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(uint8)
	if !ok {
		return 0
	}
	return value
}

// GetUint16 returns the value associated with the key as a uint16.
func (s *settings) GetUint16(key string) uint16 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(uint16)
	if !ok {
		return 0
	}
	return value
}

// GetUint32 returns the value associated with the key as a uint32.
func (s *settings) GetUint32(key string) uint32 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(uint32)
	if !ok {
		return 0
	}
	return value
}

// GetUint64 returns the value associated with the key as a uint64.
func (s *settings) GetUint64(key string) uint64 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(uint64)
	if !ok {
		return 0
	}
	return value
}

// GetFloat32 returns the value associated with the key as a float32.
func (s *settings) GetFloat32(key string) float32 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(float32)
	if !ok {
		return 0
	}
	return value
}

// GetFloat64 returns the value associated with the key as a float64.
func (s *settings) GetFloat64(key string) float64 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(float64)
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

	value, ok := v.(string)
	if !ok {
		return ""
	}
	return value
}

// GetValue returns the value associated with the key as a interface{}.
func (s *settings) GetValue(key string) interface{} {
	s.RLock()
	defer s.RUnlock()

	return s.data[key]
}
