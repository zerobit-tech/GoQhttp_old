package mssqlserver

var SPParamStringTypes map[string]string = map[string]string{

	"CHAR":     "STRING",
	"VARCHAR":  "STRING",
	"TEXT":     "STRING",
	"NCHAR":    "STRING",
	"NVARCHAR": "STRING",
	"NTEXT":    "STRING",
}

var SPParamNumericTypes map[string]string = map[string]string{

	"BIGINT":     "NUMERIC",
	"BIT":        "NUMERIC",
	"DECIMAL":    "NUMERIC",
	"INT":        "NUMERIC",
	"MONEY":      "NUMERIC",
	"NUMERIC":    "NUMERIC",
	"SMALLINT":   "NUMERIC",
	"SMALLMONEY": "NUMERIC",
	"TINYINT":    "NUMERIC",
	"FLOAT":      "NUMERIC",
	"REAL":       "NUMERIC",
}

// without decimal point
var SPParamIntegerTypes map[string]string = map[string]string{

	"SMALLINT": "NUMERIC",
	"BIGINT":   "NUMERIC",
	"INT":      "NUMERIC",
	"TINYINT":  "NUMERIC",
}

var SPParamDateTypes map[string]string = map[string]string{

	"TIME":           "DATE",
	"DATE":           "DATE",
	"DATETIME2":      "DATE",
	"DATETIME":       "DATE",
	"DATETIMEOFFSET": "DATE",
	"SMALLDATETIME":  "DATE",

	//"DISTINCT"
	//"ARRAY"

}
