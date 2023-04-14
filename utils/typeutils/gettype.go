package typeutils

import "reflect"

func TypeOfValue(value any) reflect.Kind {
	return reflect.ValueOf(value).Kind()
}

func IsBoolean(value any) bool {
	return reflect.ValueOf(value).Kind() == reflect.Bool
}
