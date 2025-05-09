package structutil

import (
	"reflect"
	"strings"
)

func DiffUpdateMap(oldObj, newObj any, ignoreFields ...string) map[string]any {
	ignore := make(map[string]struct{})
	for _, f := range ignoreFields {
		ignore[f] = struct{}{}
	}

	changes := make(map[string]any)

	oldVal := reflect.ValueOf(oldObj)
	newVal := reflect.ValueOf(newObj)
	oldType := reflect.TypeOf(oldObj)

	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
		newVal = newVal.Elem()
		oldType = oldType.Elem()
	}

	for i := 0; i < oldType.NumField(); i++ {
		field := oldType.Field(i)
		oldField := oldVal.Field(i)
		newField := newVal.Field(i)

		if !oldField.CanInterface() {
			continue
		}

		// 值相同或在忽略列表中就跳过
		gormTag := field.Tag.Get("gorm")
		col := parseGormColumn(gormTag, field.Name)
		if _, skip := ignore[col]; skip {
			continue
		}

		if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
			changes[col] = newField.Interface()
		}
	}

	return changes
}

func parseGormColumn(tag string, fallback string) string {
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return fallback
}
