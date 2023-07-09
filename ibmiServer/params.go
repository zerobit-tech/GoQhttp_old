package ibmiServer

import (
	"fmt"
	"strings"

	"github.com/onlysumitg/GoQhttp/go_ibm_db"
	"github.com/onlysumitg/GoQhttp/internal/storedProc"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *IBMiServer) GetDefaultValue(p *storedProc.StoredProcParamter) string {
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

func getSpecialRegisterValue(s *IBMiServer, name string) string {
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
