package typeutils

import "reflect"

func IsMap(value any) bool {
	return reflect.ValueOf(value).Kind() == reflect.Map
}
