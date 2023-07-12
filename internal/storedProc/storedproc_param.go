package storedProc

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/onlysumitg/GoQhttp/utils/floatutils"
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

	switch strings.ToUpper(p.Datatype) {

	case "DECFLOAT":

		x, ok := (*v).([]byte)
		if ok {
			return strconv.ParseFloat(string(x), 64)
		}

	case "NUMERIC", "DECIMAL":

		x, ok := (*v).(float64)
		if ok {

			if floatutils.AlmostEquals(x, 0, 0.00000000000000001) {
				return 0, nil
			}
			return x, nil
		}

	case "ROWID":

		x, ok := (*v).([]byte)
		if ok {
			return strconv.Atoi(string(x))
		}

	case "DATE":

		x, ok := (*v).(string)
		if ok {
			if x == "-0001-11-30" {
				return "0001-01-01", nil
			}

		}
	case "TIMESTAMP":

		x, ok := (*v).(string)
		if ok {
			if x == "-0001-11-30 00:00:00.000000" {
				return "0001-01-01 00:00:00.000000", nil
			}

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
