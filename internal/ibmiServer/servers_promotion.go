package ibmiServer

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/onlysumitg/godbc"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
)

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) listAutoPromotion() ([]*storedProc.PromotionRecord, error) {
	promotionRecords := make([]*storedProc.PromotionRecord, 0)
	if strings.TrimSpace(s.AutoPromotePrefix) != "" && strings.TrimSpace(s.ConfigFileLib) != "" {
		prefixToCheck := strings.ToUpper(strings.TrimSpace(s.AutoPromotePrefix)) + "%"
		if s.LastAutoPromoteDate == "" {
			s.LastAutoPromoteDate = time.Now().Format(godbc.TimestampFormat)
		}

		sqlToUse := fmt.Sprintf("select upper(trim(SPECIFIC_NAME)) from qsys2.sysprocs where upper(SPECIFIC_NAME) like '%s' and SPECIFIC_SCHEMA='%s' and ROUTINE_CREATED >= '%s'", strings.ToUpper(prefixToCheck), strings.ToUpper(s.ConfigFileLib), s.LastAutoPromoteDate)
		conn, err := s.GetSingleConnection()
		if err != nil {

			return promotionRecords, err
		}
		defer conn.Close()

		rows, err := conn.Query(sqlToUse)

		if err != nil {
			// var odbcError *odbc.Error

			// if errors.As(err, &odbcError) {
			// 	s.UpdateAfterError(odbcError)
			// }
			return promotionRecords, err
		}
		defer rows.Close()

		for rows.Next() {
			spName := ""
			err := rows.Scan(&spName)
			if err == nil {

				rcd := &storedProc.PromotionRecord{}
				brokenSPName := strings.Split(spName, "_")
				if len(brokenSPName) != 3 {
					log.Println("Auto promotion record skipped for SP(Name format is not correct):", spName)
				} else {
					rcd.Status = "P"
					rcd.Rowid = ""
					rcd.Endpoint = brokenSPName[2]
					rcd.Action = "I"
					rcd.Storedproc = spName
					rcd.Storedproclib = s.ConfigFileLib
					rcd.Httpmethod = brokenSPName[1]
					rcd.UseSpecificName = "Y"
					promotionRecords = append(promotionRecords, rcd)
				}

			}
		}
	}

	return promotionRecords, nil

}
