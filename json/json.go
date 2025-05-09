package json

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var json = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
	UseNumber:              true,
}.Froze()

var (
	Marshal             = json.Marshal
	Unmarshal           = json.Unmarshal
	MarshalToString     = json.MarshalToString
	UnmarshalFromString = json.UnmarshalFromString
	NewEncoder          = json.NewEncoder
	NewDecoder          = json.NewDecoder
)

func BindValue(v interface{}, dst interface{}) error {
	vByte, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "bind value, marshal v error")
	}
	if err = json.Unmarshal(vByte, dst); err != nil {
		return errors.Wrap(err, "bind value, unmarshal to dst error")
	}
	return nil
}
