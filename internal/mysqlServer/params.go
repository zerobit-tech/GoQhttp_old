package mysqlserver

import (
	"fmt"
	"strings"

	"github.com/onlysumitg/GoQhttp/go_ibm_db"
	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/utils/stringutils"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MySqlServer) GetDefaultValue(p *storedProc.StoredProcParamter) string {
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

func getSpecialRegisterValue(s *MySqlServer, name string) string {
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
// parameter
// -----------------------------------------------------------------
func getParameterofType(p *storedProc.StoredProcParamter) *any {
	var x any
	switch strings.ToUpper(p.Datatype) {

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
func parameterIsString(p *storedProc.StoredProcParamter) bool {
	_, found := go_ibm_db.SPParamStringTypes[p.Datatype]

	//_, found2 := go_ibm_db.SPParamDateTypes[p.Datatype]

	return found //|| found2
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func parameterIsNumeric(p *storedProc.StoredProcParamter) bool {
	_, found := go_ibm_db.SPParamNumericTypes[p.Datatype]
	return found
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func parameterIsInt(p *storedProc.StoredProcParamter) bool {
	_, found := go_ibm_db.SPParamIntegerTypes[p.Datatype]
	return found
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func parameterNeedQuote(p *storedProc.StoredProcParamter, value string) bool {
	// if go_ibm_db.IsSepecialRegister(value) {
	// 	return false
	// }

	if value == "NULL" {
		return false
	}
	return true
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func parameterHasValidValue(p *storedProc.StoredProcParamter, val any) bool {

	if parameterIsInt(p) {
		return stringutils.IsNumericWithOutDecimal(stringutils.AsString(val))
	}

	if parameterIsNumeric(p) {
		return stringutils.IsNumeric(stringutils.AsString(val))
	}
	return true
}
