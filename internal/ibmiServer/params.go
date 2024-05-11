package ibmiServer

import (
	"fmt"
	"strings"

	"github.com/onlysumitg/godbc"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *Server) getDefaultValue(p *storedProc.StoredProcParamter) string {
	if p.DefaultValue.Valid {

		if godbc.IsSepecialRegister(p.DefaultValue.String) {
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

func getSpecialRegisterValue(s *Server, name string) string {
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
	switch p.Datatype {
	case "DECFLOAT":
		var decfloac float64
		x = &decfloac
		return &x
	case "ROWID":
		var r godbc.ROWID
		x = &r
		return &x
	}

	return &x
}
