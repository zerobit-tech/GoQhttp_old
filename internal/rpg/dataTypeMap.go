package rpg

import "strconv"

// ------------------------------------------------------------
//
// ------------------------------------------------------------
var DataTypeMap map[string]string = map[string]string{

	"INTEGER 3i0":  "3i0",
	"INTEGER 5i0":  "5i0",
	"INTEGER 10i0": "10i0",
	"INTEGER 20i0": "20i0",

	"UNSIGNED INTEGER 3u0":  "3u0",
	"UNSIGNED INTEGER 5u0":  "5u0",
	"UNSIGNED INTEGER 10u0": "10u0",
	"UNSIGNED INTEGER 20u0": "20u0",

	"PACKED":   "%dp%d",
	"ZONED":    "%ds%d",
	"FLOAT 4f": "4f2",
	"FLOAT 8f": "8f2",

	"VARYING ALPHANUMERIC/CHAR": "%da", //<data type='32a' varying='on'/>
	"ALPHANUMERIC/CHAR":         "%da",
	"TIME":                      "8A",
	"TIMESTAMP":                 "26A",
	"DATE":                      "10A",
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
var dataTypeWithLength map[string]bool = map[string]bool{

	"ALPHANUMERIC/CHAR":         true,
	"VARYING ALPHANUMERIC/CHAR": true, //<data type='32a' varying='on'/>

	"PACKED": true,

	"ZONED": true,
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
var dataTypeWithDecimal map[string]bool = map[string]bool{

	"PACKED": true,

	"ZONED": true,
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
var dataTypeNumeric map[string]bool = map[string]bool{

	"INTEGER 3i0":  true,
	"INTEGER 5i0":  true,
	"INTEGER 10i0": true,
	"INTEGER 20i0": true,

	"UNSIGNED INTEGER 3u0":  true,
	"UNSIGNED INTEGER 5u0":  true,
	"UNSIGNED INTEGER 10u0": true,
	"UNSIGNED INTEGER 20u0": true,

	"PACKED": true,

	"ZONED":    true,
	"FLOAT 4f": true,
	"FLOAT 8f": true,
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func DataTypeNeedDecimalValue(dataType string) bool {
	_, found := dataTypeWithDecimal[dataType]

	return found
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func DataTypeNeedLength(dataType string) bool {
	_, found := dataTypeWithLength[dataType]

	return found
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
var DataTypeValidator map[string]func(string, int, int) bool = map[string]func(string, int, int) bool{

	"INTEGER 3i0":  intValidator,
	"INTEGER 5i0":  intValidator,
	"INTEGER 10i0": intValidator,
	"INTEGER 20i0": intValidator,

	"UNSIGNED INTEGER 3u0":  uintValidator,
	"UNSIGNED INTEGER 5u0":  uintValidator,
	"UNSIGNED INTEGER 10u0": uintValidator,
	"UNSIGNED INTEGER 20u0": uintValidator,

	"PACKED":   floatValidator,
	"ZONED":    floatValidator,
	"FLOAT 4f": floatValidator,
	"FLOAT 8f": floatValidator,

	"VARYING ALPHANUMERIC/CHAR": charValidator, //<data type='32a' varying='on'/>
	"ALPHANUMERIC/CHAR":         charValidator,
	"TIME":                      timeValidator,
	"TIMESTAMP":                 dateValidator,
	"DATE":                      timestampValidator,
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
var DataTypeSpecialText map[string]string = map[string]string{

	"VARYING ALPHANUMERIC/CHAR": "varying='on'",
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func intValidator(v string, length, decimal int) bool {
	_, err := strconv.Atoi(v)
	if err != nil {
		return false // error
	}
	return true
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func uintValidator(v string, length, decimal int) bool {
	i, err := strconv.Atoi(v)
	if err != nil {
		return false // error
	}

	if i < 0 {
		return false // error
	}
	return true
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func floatValidator(v string, length, decimal int) bool {
	_, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return false // error
	}
	return true
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func charValidator(v string, length, decimal int) bool {
	if len(v) > length {
		return false
	}

	return true
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func timeValidator(v string, length, decimal int) bool {
	if len(v) > 8 {
		return false
	}

	return true
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func dateValidator(v string, length, decimal int) bool {
	if len(v) > 10 {
		return false
	}

	return true
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func timestampValidator(v string, length, decimal int) bool {
	if len(v) > 26 {
		return false
	}

	return true
}
