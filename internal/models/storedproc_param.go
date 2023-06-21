package models

import (
	"database/sql"
	"fmt"
	"reflect"
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
		fmt.Println("asString(v)", asString(v))

		x, ok := (*v).([]byte)
		if ok {
			return strconv.ParseFloat(string(x), 64)
		}
	case "ROWID":
		fmt.Println("asString(v)", asString(v))

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
		return time.Parse(go_ibm_db.TimeFormat, asString(v))

	case "DATE":
		return time.Parse(go_ibm_db.DateFormat, asString(v))

	case "TIMESTAMP":
		return time.Parse(go_ibm_db.TimestampFormat, asString(v))

	case "SMALLINT", "INTEGER", "BIGINT", "ROWID":
		if v == nil {
			var x int = 0

			return x, nil
		}
		return strconv.Atoi(asString(v))

	case "DECIMAL", "NUMERIC", "DECFLOAT", "DOUBLE PRECISION", "REAL":
		return strconv.ParseFloat(asString(v), 64)

	}

	return v, nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) GetDefaultValue(s Server) string {
	if p.DefaultValue.Valid {

		if go_ibm_db.IsSepecialRegister(p.DefaultValue.String) {
			return getSpecialRegisterValue(s, p.DefaultValue.String)
		}
		d := strings.ReplaceAll(p.DefaultValue.String, " ", "")
		if d == "''" {
			return ""
		}

		return (p.DefaultValue.String)
	}
	return ""
}

func getSpecialRegisterValue(s Server, name string) string {
	sqlToUse := fmt.Sprintf("values(%s)", name)
	conn, err := s.GetConnection()

	var valToUse string
	if err != nil {

		return ""
	}

	row := conn.QueryRow(sqlToUse)
	err = row.Scan(&valToUse)
	if err == nil {
		return valToUse
	}
	return ""

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
		return stringutils.IsNumericWithOutDecimal(asString(val))
	}

	if p.IsNumeric() {
		return stringutils.IsNumeric(asString(val))
	}
	return true
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

func asString(src interface{}) string {

	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case *[]byte:
		return string(*v)
	}
	rv := reflect.ValueOf(src)
	//fmt.Println("rv.Kind()", rv.Kind())
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())

	}
	return fmt.Sprintf("%v", src)
}
