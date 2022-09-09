package types

import (
	cherryMapStructure "github.com/cherry-game/cherry/extend/mapstructure"
	"github.com/spf13/cast"
	"reflect"
)

type (
	//I32I32  int32&int32 key,value
	I32I32 struct {
		Key   int32 // key
		Value int32 // value
	}
)

func (I32I32) Type() reflect.Type {
	return reflect.TypeOf(I32I32{})
}

func (p *I32I32) Hook() cherryMapStructure.DecodeHookFuncType {
	return func(_ reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if t == p.Type() {
			return p.parseStruct(data), nil
		}
		return data, nil
	}
}

func (p *I32I32) parseStruct(values interface{}) I32I32 {
	var rsp I32I32

	if values == nil {
		return rsp
	}

	valuesSlice := cast.ToSlice(values)
	if valuesSlice == nil {
		return rsp
	}

	if len(valuesSlice) == 2 {
		k, starErr := cast.ToInt32E(valuesSlice[0])
		v, numErr := cast.ToInt32E(valuesSlice[1])
		if starErr == nil && numErr == nil {
			rsp.Key = k
			rsp.Value = v
		}
	}

	return rsp
}
