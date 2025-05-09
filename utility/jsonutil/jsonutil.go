package jsonutil

import (
	"fmt"
	"foundation-go/json"
	"foundation-go/utility/reflectutil"
	"foundation-go/utility/stringutil"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type StringValue string

func (s *StringValue) UnmarshalJSON(b []byte) (err error) {
	strVal := stringutil.FromBytes(b)
	// 判断是 string
	if strings.HasPrefix(strVal, `"`) && strings.HasSuffix(strVal, `"`) {
		err = json.Unmarshal(b, &strVal)
	}
	*s = StringValue(strVal)
	return err
}

// only receive map and struct
// other types will return nil
func ToStringMapString(obj interface{}) (map[string]string, error) {
	switch v := obj.(type) {
	case nil:
		return nil, nil
	case map[string]string:
		return v, nil
	}

	bytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	result, temp := map[string]string{}, map[string]StringValue{}
	err = json.Unmarshal(bytes, &temp)

	if err != nil {
		return result, err
	}

	for key, val := range temp {
		result[key] = string(val)
	}

	return result, err
}

// only receive map and struct
// other types will return nil
func ToStringMapInterface(obj interface{}) (map[string]interface{}, error) {
	switch v := obj.(type) {
	case nil:
		return nil, nil
	case map[string]interface{}:
		return v, nil
	}

	if !IsJSONObject(obj) {
		format := "json object should be map or struct, got a :%s"
		return nil, fmt.Errorf(format, reflectutil.TypeOf(obj))
	}

	jsonObj := map[string]interface{}{}
	jsonStr, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonStr, &jsonObj)
	return jsonObj, err
}

// IsJSONObject If the type of obj is map or struct, return true
func IsJSONObject(obj interface{}) bool {
	switch reflectutil.TypeOf(obj) {
	case reflect.Map, reflect.Struct:
		return true
	default:
		return false
	}
}

func ToJSON(val interface{}) string {
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return "{}"
	}
	return stringutil.FromBytes(jsonBytes)
}

func BindValue(val, dest any) error {
	bytes, err := json.Marshal(val)
	if err != nil {
		return errors.Wrap(err, "BindValue marshal val error")
	}
	if err := json.Unmarshal(bytes, &dest); err != nil {
		return errors.Wrap(err, "bind value error")
	}
	return nil
}
