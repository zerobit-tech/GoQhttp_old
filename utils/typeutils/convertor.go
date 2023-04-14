package typeutils

import (
	"reflect"
	"strconv"
	"strings"
)

func ConvertToType(value, dataType string) any {
	switch strings.ToUpper(dataType) {
	case "BOOL": // without quotes
		return GetBoolVal(value)

	case "FLOAT64":
		return GetFloatVal(value)

	case "INT":
		return GetIntVal(value)
	default:
		return value

	}
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func GetBoolVal(valueToUse any) bool {
	var boolVal bool = false
	if reflect.ValueOf(valueToUse).Kind() == reflect.String {
		boolValX, err := strconv.ParseBool(valueToUse.(string))

		if err == nil {
			boolVal = boolValX
		}
	} else {
		boolValX, ok := valueToUse.(bool)
		if !ok {

		} else {
			boolVal = boolValX

		}
	}
	return boolVal
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func GetFloatVal(valueToUse any) float64 {
	var floatVal float64 = 0
	if reflect.ValueOf(valueToUse).Kind() == reflect.String {
		floatValX, err := strconv.ParseFloat(valueToUse.(string), 64)

		if err == nil {
			floatVal = floatValX
		}
	} else {
		floatValX, ok := valueToUse.(float64)
		if !ok {

		} else {
			floatVal = floatValX

		}
	}
	return floatVal
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func GetIntVal(valueToUse any) int {
	var intval int = 0
	if reflect.ValueOf(valueToUse).Kind() == reflect.String {
		intvalX, err := strconv.Atoi(valueToUse.(string))

		if err == nil {
			intval = intvalX
		}
	} else {
		intvalX, ok := valueToUse.(int)
		if !ok {

		} else {
			intval = intvalX

		}
	}
	return intval
}
