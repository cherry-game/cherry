package cherryDataConfig

import (
	"reflect"
	"strings"
)

var (
	modelMaps = make(map[string]interface{})
)

func Add(table string, value interface{}) {
	if value == nil {
		return
	}
	key := strings.ToLower(table)
	modelMaps[key] = value
}

func List(table string) interface{} {
	return modelMaps[strings.ToLower(table)]
}

func EqualOne(table string, key string, value interface{}) interface{} {
	list := List(table)
	if list == nil {
		return nil
	}

	if reflect.TypeOf(list).Kind() != reflect.Slice {
		return nil
	}

	s := reflect.ValueOf(list)
	for i := 0; i < s.Len(); i++ {
		ele := s.Index(i)
		k := ele.Elem().FieldByName(key)

		if !k.IsValid() {
			return nil
		}

		if compare(k, value) {
			return ele.Interface().(IConfigModel)
		}
	}
	return nil
}

func EqualList(table string, key string, value interface{}) interface{} {
	models := List(table).([]IConfigModel)
	if models == nil {
		//errors.NewFormat("table=%s not found. key=%s, value=%s", table, key, value)
		return nil
	}

	var out []IConfigModel

	for _, model := range models {
		vf := reflect.ValueOf(model)
		v := vf.Elem().FieldByName(key)
		if !v.IsValid() {
			return nil
		}

		if v.Kind() != reflect.ValueOf(value).Kind() {
			return nil
		}

		if compare(v, value) {
			out = append(out, model)
		}
	}

	return out
}

func compare(c1 reflect.Value, c2 interface{}) bool {
	if c1.Kind() != reflect.ValueOf(c2).Kind() {
		return false
	}

	switch c2.(type) {
	case int:
		if c1.Int() == int64(c2.(int)) {
			return true
		}
	case string:
		if c1.String() == c2.(string) {
			return true
		}
	}

	return false
}

func Or(table string, params map[string]interface{}, out interface{}) error {
	return nil
}
