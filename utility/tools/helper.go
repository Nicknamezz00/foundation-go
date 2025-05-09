package tools

import (
	"foundation-go/json"
	"reflect"
)

func AssignNullable[T, M any](src *T, dst **M) {
	if src == nil {
		return
	}
	// 非nil看类型是否匹配
	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()
	// 如果类型直接兼容（同类型），就直接赋值
	if srcVal.Type().AssignableTo(dstVal.Type()) {
		dstVal.Set(srcVal)
		return
	}
	// 特殊情况：struct → string
	if (srcVal.Kind() == reflect.Struct || srcVal.Kind() == reflect.Map) && dstVal.Kind() == reflect.String {
		s, err := json.MarshalToString(src)
		if err == nil {
			dstVal.Set(reflect.ValueOf(s))
		} else {
			panic(err)
		}
	}
}
