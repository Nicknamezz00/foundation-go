package reflectutil

import "reflect"

func TypeOf(v interface{}) reflect.Kind {
	val := reflect.Indirect(reflect.ValueOf(v))
	return val.Kind()
}

func ValueOf(ptr interface{}) interface{} {
	return reflect.ValueOf(ptr).Elem().Interface()
}
