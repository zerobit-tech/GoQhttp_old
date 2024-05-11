package httputils

import (
	"net/http"
	"reflect"
)

func IsValidHttpCode(code int) (valid bool, message string) {
	message = http.StatusText(code)

	valid = (message != "")

	return
}

func GetValidHttpCode(v any) (code int, message string) {
	intval, ok := 0, false

	qhttp_status_code := 0
	qhttp_status_message :=""
	switch reflect.ValueOf(v).Kind() {
	case reflect.Int32:
		if intval32, ok2 := (v).(int32); ok2 {
			intval = int(intval32)
			ok = ok2
		}
	case reflect.Int64:
		if intval64, ok2 := (v).(int64); ok2 {
			intval = int(intval64)
			ok = ok2
		}
	case reflect.Int16:
		if intval16, ok2 := (v).(int16); ok2 {
			intval = int(intval16)
			ok = ok2
		}
	case reflect.Int8:
		if intval8, ok2 := (v).(int8); ok2 {
			intval = int(intval8)
			ok = ok2
		}
	default:
		intval, ok = (v).(int)
	}

	if ok {
		validCode, message := IsValidHttpCode(int(intval))
		if validCode {
			qhttp_status_code = int(intval)
			qhttp_status_message = message
 
		}
	}

	return qhttp_status_code,qhttp_status_message
}
