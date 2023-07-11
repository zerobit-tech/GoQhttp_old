package storedProc

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/onlysumitg/GoQhttp/utils/stringutils"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

type StoredProcParamter struct {
	Position           int
	Mode               string // in out inout
	Name               string
	Alias              string // alias to original name
	Datatype           string //todo list all avaialble data types
	Scale              int
	Precision          int
	MaxLength          int
	DefaultValue       sql.NullString
	GlobalVariableName string
	CreateStatement    string
	DropStatement      string
	GivenValue         string
	OutValue           string
	validForCall       bool `json:"-" db:"-" form:"-"`
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) GetNameToUse() string {
	paramNameToUse := p.Name

	if p.Alias != "" {
		paramNameToUse = p.Alias

	}
	return paramNameToUse
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) GetDefaultValueX() string {
	if p.DefaultValue.Valid {
		d := strings.ReplaceAll(p.DefaultValue.String, " ", "")
		if d == "''" {
			return ""
		}

		return p.DefaultValue.String
	}
	return ""
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) ConvertOUTVarToType(v *any) (any, error) {
	if p.Mode != "OUT" {
		return v, nil
	}

	switch p.Datatype {

	case "DECFLOAT":
		fmt.Println("asString(v)", stringutils.AsString(v))

		x, ok := (*v).([]byte)
		if ok {
			return strconv.ParseFloat(string(x), 64)
		}
	case "ROWID":
		fmt.Println("asString(v)", stringutils.AsString(v))

		x, ok := (*v).([]byte)
		if ok {
			return strconv.Atoi(string(x))
		}
	}

	return v, nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) ConvertToType(v any) (any, error) {

	var TimeFormat string = "15:04:05"
	var DateFormat string = "2006-01-02"
	var TimestampFormat string = "2006-01-02 15:04:05.000000"

	switch p.Datatype {
	case "TIME":
		return time.Parse(TimeFormat, stringutils.AsString(v))

	case "DATE":
		return time.Parse(DateFormat, stringutils.AsString(v))

	case "TIMESTAMP":
		return time.Parse(TimestampFormat, stringutils.AsString(v))

	case "SMALLINT", "INTEGER", "BIGINT", "ROWID":
		if v == nil {
			var x int = 0

			return x, nil
		}
		return strconv.Atoi(stringutils.AsString(v))

	case "DECIMAL", "NUMERIC", "DECFLOAT", "DOUBLE PRECISION", "REAL":
		return strconv.ParseFloat(stringutils.AsString(v), 64)

	}

	return v, nil
}
