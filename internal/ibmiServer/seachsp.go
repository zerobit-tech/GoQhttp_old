package ibmiServer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
)

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) SearchSP(libname, spname string) ([]*storedProc.StoredProc, error) {

	spList := make([]*storedProc.StoredProc, 0)
	if strings.TrimSpace(libname) == "" {
		libname = "*LIBL"
	}
	if strings.TrimSpace(libname) == "" || strings.TrimSpace(spname) == "" {
		return spList, errors.New("Both Library and Stored procedure name are required")
	}

	baseQuery := "select trim(SPECIFIC_SCHEMA), trim(SPECIFIC_NAME),trim(ROUTINE_SCHEMA),trim(ROUTINE_NAME),    ROUTINE_CREATED from qsys2.sysprocs"

	spSearch := fmt.Sprintf("(SPECIFIC_NAME like '%s%%' or ROUTINE_NAME like '%s%%')", strings.ToUpper(spname), strings.ToUpper(spname))

	libSearch := fmt.Sprintf("(SPECIFIC_SCHEMA='%s' or ROUTINE_SCHEMA='%s')", strings.ToUpper(libname), strings.ToUpper(libname))

	if strings.EqualFold(libname, "*LIBL") {

		libSearch = "ROUTINE_SCHEMA in (SELECT SYSTEM_SCHEMA_NAME FROM QSYS2.LIBRARY_LIST_INFO order by ORDINAL_POSITION)"
	}
	sqlToUse := fmt.Sprintf("%s where %s and %s order by ROUTINE_CREATED desc limit 50", baseQuery, spSearch, libSearch)

	conn, err := s.GetSingleConnection()
	if err != nil {
		return spList, err
	}
	defer conn.Close()

	rows, err := conn.Query(sqlToUse)

	if err != nil {

		return spList, err
	}
	defer rows.Close()

	for rows.Next() {
		spE := &storedProc.StoredProc{}

		err = rows.Scan(&spE.SpecificLib, &spE.SpecificName, &spE.Lib, &spE.Name, &spE.Modified)
		if err != nil {
			break
		} else {
			spList = append(spList, spE)
		}

	}

	return spList, nil

}
