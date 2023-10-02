package responseprocessor

import (
	"regexp"
	"strconv"
	"strings"
)

var dataTypeNumericByCode map[string]func(string) any = map[string]func(string) any{

	"3I0":  ToInt,
	"5I0":  ToInt,
	"10I0": ToInt,
	"20I0": ToInt,

	"3U0":  ToInt,
	"5U0":  ToInt,
	"10U0": ToInt,
	"20U0": ToInt,

	"PACKED": ToFloat,
	"ZONED":  ToFloat,

	"4F2": ToFloat,
	"8F2": ToFloat,
}

func isPackedOrZoned(dt string) bool {
	var re = regexp.MustCompile(`(?m)\d*[sp]\d`)
	return re.MatchString(dt)
}

func toGoVal(dt string, val string) any {

	convertor, found := dataTypeNumericByCode[strings.ToUpper(dt)]
	if found {
		return convertor(val)
	} else {
		if isPackedOrZoned(dt) {
			return ToFloat(val)
		}
	}

	return val
}

func ToInt(v string) any {
	i, err := strconv.Atoi(v)
	if err == nil {
		return i
	}

	return v
}

func ToFloat(v string) any {
	i, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return i
	}

	return v
}
