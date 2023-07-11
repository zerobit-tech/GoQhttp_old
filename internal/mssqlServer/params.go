package mssqlserver

import (
	"fmt"
	"strings"

	"github.com/onlysumitg/GoQhttp/internal/storedProc"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MSSqlServer) GetDefaultValue(p *storedProc.StoredProcParamter) string {
	if p.DefaultValue.Valid {

		// if go_ibm_db.IsSepecialRegister(p.DefaultValue.String) {
		// 	return getSpecialRegisterValue(s, p.DefaultValue.String)
		// }
		d := strings.ReplaceAll(p.DefaultValue.String, " ", "")
		if d == "''" {
			return ""
		}

		return (p.DefaultValue.String)
	}
	return ""
}

func getSpecialRegisterValue(s *MSSqlServer, name string) string {
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
func parameterIsString(p *storedProc.StoredProcParamter) bool {
	_, found := SPParamStringTypes[p.Datatype]

	//_, found2 := go_ibm_db.SPParamDateTypes[p.Datatype]

	return found //|| found2
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func parameterIsNumeric(p *storedProc.StoredProcParamter) bool {
	_, found := SPParamNumericTypes[p.Datatype]
	return found
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func parameterIsInt(p *storedProc.StoredProcParamter) bool {
	_, found := SPParamIntegerTypes[p.Datatype]
	return found
}

// // -----------------------------------------------------------------
// //
// // -----------------------------------------------------------------
// func parameterNeedQuote(p *storedProc.StoredProcParamter, value string) bool {
// 	if go_ibm_db.IsSepecialRegister(value) {
// 		return false
// 	}

// 	if value == "NULL" {
// 		return false
// 	}
// 	return true
// }

// // -----------------------------------------------------------------
// //
// // -----------------------------------------------------------------
// func parameterHasValidValue(p *storedProc.StoredProcParamter, val any) bool {

// 	if parameterIsInt(p) {
// 		return stringutils.IsNumericWithOutDecimal(stringutils.AsString(val))
// 	}

// 	if parameterIsNumeric(p) {
// 		return stringutils.IsNumeric(stringutils.AsString(val))
// 	}
// 	return true
// }
