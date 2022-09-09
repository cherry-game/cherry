package types

import (
	"encoding/json"
	cherryMapStructure "github.com/cherry-game/cherry/extend/mapstructure"
	"github.com/spf13/cast"
	"reflect"
)

type I32I64Map map[int32]int64

func NewI32I64Map() I32I64Map {
	return make(map[int32]int64)
}

func (I32I64Map) Type() reflect.Type {
	return reflect.TypeOf(I32I64Map{})
}

func (p *I32I64Map) Hook() cherryMapStructure.DecodeHookFuncType {
	return func(_ reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if t == p.Type() {
			return p.ToMap(data), nil
		}
		return data, nil
	}
}

func (p *I32I64Map) ToMap(values interface{}) map[int32]int64 {
	var maps = make(map[int32]int64)
	if values == nil {
		return maps
	}

	valueSlice := cast.ToSlice(values)
	if valueSlice == nil {
		return maps
	}

	if len(valueSlice) == 2 {
		k, kErr := cast.ToInt32E(valueSlice[0])
		v, vErr := cast.ToInt64E(valueSlice[1])
		if kErr == nil && vErr == nil {
			maps[k] = v
			return maps
		}
	}

	for _, value := range valueSlice {
		result, found := value.([]interface{})
		if found == false {
			break
		}

		if len(result) >= 2 {
			k := cast.ToInt32(result[0])
			v := cast.ToInt64(result[1])
			maps[k] = v
		}
	}

	return maps
}

func (p I32I64Map) ReadString(data string) {
	var jsonObject interface{}
	err := json.Unmarshal([]byte(data), &jsonObject)
	if err != nil {
		return
	}

	resultMap := p.ToMap(jsonObject)
	for k, v := range resultMap {
		p[k] = v
	}
}

func (p I32I64Map) Decrease(key int32, decreaseValue int64) (int64, bool) {
	if decreaseValue < 1 {
		return 0, false
	}

	value, _ := p[key]
	if value < decreaseValue {
		return 0, false
	}

	p[key] = value - decreaseValue

	return p[key], true
}

func (p I32I64Map) Add(key int32, addValue int64) (int64, bool) {
	if addValue < 1 {
		return 0, false
	}

	value, _ := p[key]
	p[key] = value + addValue

	return p[key], true
}

func (p I32I64Map) Get(key int32) (int64, bool) {
	value, found := p[key]
	if found {
		return value, true
	}
	return 0, false
}

func (p I32I64Map) Set(key int32, value int64) {
	p[key] = value
}

func (p I32I64Map) ContainKey(key int32) bool {
	_, found := p[key]
	return found
}
