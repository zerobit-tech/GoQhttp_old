package storedProc

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/onlysumitg/GoQhttp/go_ibm_db"
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
func (p *StoredProcParamter) GetofType() *any {
	var x any
	switch p.Datatype {
	case "DECFLOAT":
		var decfloac float64
		x = &decfloac
		return &x
	case "ROWID":
		var r go_ibm_db.ROWID
		x = &r
		return &x
	}

	return &x
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
	switch p.Datatype {
	case "TIME":
		return time.Parse(go_ibm_db.TimeFormat, stringutils.AsString(v))

	case "DATE":
		return time.Parse(go_ibm_db.DateFormat, stringutils.AsString(v))

	case "TIMESTAMP":
		return time.Parse(go_ibm_db.TimestampFormat, stringutils.AsString(v))

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



// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) IsString() bool {
	_, found := go_ibm_db.SPParamStringTypes[p.Datatype]

	//_, found2 := go_ibm_db.SPParamDateTypes[p.Datatype]

	return found //|| found2
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) IsNumeric() bool {
	_, found := go_ibm_db.SPParamNumericTypes[p.Datatype]
	return found
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) IsInt() bool {
	_, found := go_ibm_db.SPParamIntegerTypes[p.Datatype]
	return found
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) NeedQuote(value string) bool {
	if go_ibm_db.IsSepecialRegister(value) {
		return false
	}

	if value == "NULL" {
		return false
	}
	return true
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) HasValidValue(val any) bool {

	if p.IsInt() {
		return stringutils.IsNumericWithOutDecimal(stringutils.AsString(val))
	}

	if p.IsNumeric() {
		return stringutils.IsNumeric(stringutils.AsString(val))
	}
	return true
}
