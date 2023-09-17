package rpg

var DataTypeMap map[string]string = map[string]string{

	"INTEGER 3i0":  "3i0",
	"INTEGER 5i0":  "5i0",
	"INTEGER 10i0": "10i0",
	"INTEGER 20i0": "20i0",

	"UNSIGNED INTEGER 3u0":  "3u0",
	"UNSIGNED INTEGER 5u0":  "5u0",
	"UNSIGNED INTEGER 10u0": "10u0",
	"UNSIGNED INTEGER 20u0": "20u0",

	"ALPHANUMERIC/CHAR":         "%da",
	"VARYING ALPHANUMERIC/CHAR": "%da", //<data type='32a' varying='on'/>

	"PACKED": "%dp%d",

	"ZONED":    "%ds%d",
	"FLOAT 4f": "4f2",
	"FLOAT 8f": "8f2",
	"TIME":     "8A",

	"TIMESTAMP": "26A",
	"DATE":      "10A",
}

var dataTypeWithLength map[string]bool = map[string]bool{

	"ALPHANUMERIC/CHAR":         true,
	"VARYING ALPHANUMERIC/CHAR": true, //<data type='32a' varying='on'/>

	"PACKED": true,

	"ZONED": true,
}

var dataTypeWithDecimal map[string]bool = map[string]bool{

	"PACKED": true,

	"ZONED": true,
}

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

func DataTypeNeedDecimalValue(dataType string) bool {
	_, found := dataTypeWithDecimal[dataType]

	return found
}

func DataTypeNeedLength(dataType string) bool {
	_, found := dataTypeWithLength[dataType]

	return found
}
