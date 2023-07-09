package mssqlserver

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/onlysumitg/GoQhttp/go_ibm_db"
	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/internal/validator"
)

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) PromotionRecordToStoredProc(p storedProc.PromotionRecord) *storedProc.StoredProc {
	sp := &storedProc.StoredProc{
		EndPointName: p.Endpoint,
		HttpMethod:   p.Httpmethod,
		Name:         p.Storedproc,
		Lib:          p.Storedproclib,
	}
	if p.UseSpecificName == "Y" {
		sp.UseSpecificName = true
	}

	if p.UseWithoutAuth == "Y" {
		sp.AllowWithoutAuth = true
	}
	srcd := &storedProc.ServerRecord{
		ID:   s.ID,
		Name: s.Name,
	}
	sp.DefaultServer = srcd

	return sp
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) UpdateStatusForPromotionRecord(p storedProc.PromotionRecord) {
	if p.Rowid <= 0 {
		return
	}

	updateSQL := fmt.Sprintf("update %s.%s a set status='%s' , statusmessage = '%s' where rrn(a) = %d", s.ConfigFileLib, s.ConfigFile, p.Status, p.StatusMessage, p.Rowid)
	conn, err := s.GetSingleConnection()
	defer conn.Close()
	if err != nil {
		log.Println("Error updating promotion file....", err.Error())
	}
	_, err = conn.Exec(updateSQL)
	if err != nil {
		log.Println("Error updateing promotion file.... ", err.Error())
	}
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) ListPromotion(withupdate bool) ([]*storedProc.PromotionRecord, error) {

	promotionRecords := make([]*storedProc.PromotionRecord, 0)
	if strings.TrimSpace(s.ConfigFile) != "" && strings.TrimSpace(s.ConfigFileLib) != "" {

		sqlToUse := fmt.Sprintf("select rrn(a), upper(trim(action)) , upper(trim(endpoint)), trim(storedproc), trim(storedproclib), upper(trim(httpmethod)), upper(trim(usespecificname)), upper(trim(usewithoutauth)) , upper(trim(paramalias)) from %s.%s a where status=''", s.ConfigFileLib, s.ConfigFile)

		conn, err := s.GetSingleConnection()
		defer conn.Close()
		if err != nil {

			return promotionRecords, err
		}

		rows, err := conn.Query(sqlToUse)
		defer rows.Close()

		if err != nil {
			// var odbcError *odbc.Error

			// if errors.As(err, &odbcError) {
			// 	s.UpdateAfterError(odbcError)
			// }
			return promotionRecords, err
		}

		for rows.Next() {
			rcd := &storedProc.PromotionRecord{}
			err := rows.Scan(&rcd.Rowid,
				&rcd.Action,
				&rcd.Endpoint,
				&rcd.Storedproc,
				&rcd.Storedproclib,
				&rcd.Httpmethod,
				&rcd.UseSpecificName,
				&rcd.UseWithoutAuth,
				&rcd.ParamAlias,
			)
			if err != nil {
				rcd.Status = "E"
				rcd.StatusMessage = err.Error()
			} else {
				rcd.CheckField(validator.MustBeFromList(rcd.Action, "D", "I", "R"), "ErrorMsg", "Action: Invalid value")
				rcd.CheckField(validator.NotBlank(rcd.Endpoint), "ErrorMsg", "Endpoint: This field cannot be blank")
				rcd.CheckField(validator.NotBlank(rcd.Storedproc), "ErrorMsg", "Storedproc: This field cannot be blank")
				rcd.CheckField(validator.NotBlank(rcd.Storedproclib), "ErrorMsg", "Storedproclib: This field cannot be blank")
				rcd.CheckField(validator.NotBlank(rcd.Httpmethod), "ErrorMsg", "Httpmethod: This field cannot be blank")
				rcd.CheckField(validator.MustBeFromList(rcd.Httpmethod, "GET", "POST", "PUT", "DELETE"), "ErrorMsg", "Httpmethod: Invalid value")

				if rcd.Valid() {
					rcd.Status = "P"
					rcd.StatusMessage = "processing"

				} else {
					// update table with error
					rcd.Status = "E"
					rcd.StatusMessage = rcd.Validator.FieldErrors["ErrorMsg"]
				}
			}

			rcd.BreakParamAlias()
			promotionRecords = append(promotionRecords, rcd)
		}
	}
	// if withupdate && updateSQL != "" {
	// 	_, err := conn.Exec(updateSQL)
	// 	if err != nil {
	// 		log.Println("Error updateing promotion status ", err.Error())
	// 	}
	// }

	autoP, err := s.ListAutoPromotion()
	if err == nil && len(autoP) > 0 {
		promotionRecords = append(promotionRecords, autoP...)
	}
	return promotionRecords, nil
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) ListAutoPromotion() ([]*storedProc.PromotionRecord, error) {
	promotionRecords := make([]*storedProc.PromotionRecord, 0)
	if strings.TrimSpace(s.AutoPromotePrefix) != "" && strings.TrimSpace(s.ConfigFileLib) != "" {
		prefixToCheck := strings.ToUpper(strings.TrimSpace(s.AutoPromotePrefix)) + "%"
		if s.LastAutoPromoteDate == "" {
			s.LastAutoPromoteDate = time.Now().Format(go_ibm_db.TimestampFormat)
		}

		sqlToUse := fmt.Sprintf("select upper(trim(SPECIFIC_NAME)) from qsys2.sysprocs where upper(SPECIFIC_NAME) like '%s' and SPECIFIC_SCHEMA='%s' and ROUTINE_CREATED >= '%s'", strings.ToUpper(prefixToCheck), strings.ToUpper(s.ConfigFileLib), s.LastAutoPromoteDate)
		conn, err := s.GetSingleConnection()
		defer conn.Close()
		if err != nil {

			return promotionRecords, err
		}

		rows, err := conn.Query(sqlToUse)

		defer rows.Close()

		if err != nil {
			// var odbcError *odbc.Error

			// if errors.As(err, &odbcError) {
			// 	s.UpdateAfterError(odbcError)
			// }
			return promotionRecords, err
		}

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
					rcd.Rowid = 0
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
