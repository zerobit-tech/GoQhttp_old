package models

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

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
func (p *StoredProcParamter) GetDefaultValue() []byte {
	if p.DefaultValue.Valid {

		if go_ibm_db.IsSepecialRegister(p.DefaultValue.String) {
			return go_ibm_db.GetSepecialValue(p.DefaultValue.String, nil)
		}
		d := strings.ReplaceAll(p.DefaultValue.String, " ", "")
		if d == "''" {
			return nil
		}

		return []byte(p.DefaultValue.String)
	}
	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) IsString() bool {
	_, found := go_ibm_db.SPParamStringTypes[p.Datatype]
	return found
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
