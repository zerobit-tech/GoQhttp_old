package typeutils

import "reflect"

func IsList(value any) bool {
	return reflect.ValueOf(value).Kind() == reflect.Array || reflect.ValueOf(value).Kind() == reflect.Slice
}
