package storedProc

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/onlysumitg/godbc"
	"github.com/zerobit-tech/GoQhttp/internal/validator"
	"github.com/zerobit-tech/GoQhttp/utils/floatutils"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
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
	Placement          string // query  path body
	ValidatorRegex     string
	validForCall       bool `json:"-" db:"-" form:"-"`
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) GetNameToUse(ignoreSpecials bool) string {
	paramNameToUse := p.Name

	if p.Alias != "" {
		if ignoreSpecials && strings.HasPrefix(p.Alias, "*") {
			// Nothing here
		} else {

			paramNameToUse = p.Alias
		}

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

	switch p.Datatype {
	case "TIME":
		return time.Parse(stringutils.TimeFormat, stringutils.AsString(v))

	case "DATE":
		return time.Parse(stringutils.DateFormat, stringutils.AsString(v))

	case "TIMESTAMP":
		return time.Parse(stringutils.TimestampFormat, stringutils.AsString(v))

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
	_, found := godbc.SPParamStringTypes[p.Datatype]

	//_, found2 := go_ibm_db.SPParamDateTypes[p.Datatype]

	return found //|| found2
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) IsNumeric() bool {
	_, found := godbc.SPParamNumericTypes[p.Datatype]
	return found
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) IsInt() bool {
	_, found := godbc.SPParamIntegerTypes[p.Datatype]
	return found
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) NeedQuote(value string) bool {
	if godbc.IsSepecialRegister(value) {
		return false
	}

	if value == "NULL" {
		return false
	}
	return true
}

// -----------------------------------------------------------------
//
//	To add more validation  >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) HasValidValue(val any, regexMap map[string]string) error {

	stringVal := stringutils.AsString(val)
	if p.IsString() {
		if len(stringVal) > p.MaxLength {
			return fmt.Errorf("%s: Invalid length", p.GetNameToUse(true))
		}
	}

	if p.IsInt() {
		if !stringutils.IsNumericWithOutDecimal(stringVal) {
			return fmt.Errorf("%s: Invalid integer value", p.GetNameToUse(true))

		}
	}

	if p.IsNumeric() {
		if !stringutils.IsNumeric(stringVal) {
			return fmt.Errorf("%s: Invalid numeric value", p.GetNameToUse(true))

		}
	}

	switch strings.ToUpper(p.Datatype) {
	case "TIME":
		if !stringutils.IsValidTime(stringVal) {
			return fmt.Errorf("%s: Invalid Time", p.GetNameToUse(true))

		}

	case "DATE":
		if !stringutils.IsValidDate(stringVal) {
			return fmt.Errorf("%s: Invalid Date", p.GetNameToUse(true))

		}

	case "TIMESTAMP":
		if !stringutils.IsValidTimeStamp(stringVal) {
			return fmt.Errorf("%s: Invalid TimeStamp", p.GetNameToUse(true))

		}
	}

	// final regex match
	if !p.ValidatorValueByRegex(stringVal, regexMap) {
		return fmt.Errorf("%s: Invalid value. Validation failed for %s", p.GetNameToUse(true), p.ValidatorRegex)

	}

	return nil
}

// -----------------------------------------------------------------
//
//	To add more validation  >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) CheckValidatorRegex(regexMap map[string]string) error {

	upperVal := p.ValidatorRegex

	_, found := regexMap[upperVal]
	if !found {
		return fmt.Errorf("validator %s not found", upperVal)
	}

	p.ValidatorRegex = upperVal

	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) GetValidatorRegex(regexMap map[string]string) (*regexp.Regexp, error) {

	finalRegex, found := regexMap[p.ValidatorRegex]
	if !found {
		return nil, errors.New("no validator regex defined")
	}

	return regexp.Compile(finalRegex)
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (p *StoredProcParamter) ValidatorValueByRegex(val string, regexMap map[string]string) bool {

	paramRegex := strings.ToUpper(p.ValidatorRegex)
	switch paramRegex {
	case "JSON":
		return validator.MustBeJSON(val)
	case "XML":
		return validator.MustBeXML(val)
	default:
		re, err := p.GetValidatorRegex(regexMap)
		if err != nil {
			return true // Dont stop if regex does not compile
		}

		return re.MatchString(val)

	}

}
