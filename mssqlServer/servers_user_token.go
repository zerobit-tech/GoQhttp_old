package mssqlserver

import (
	"fmt"
	"log"
	"strings"

	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/internal/validator"
)

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) UpdateStatusUserTokenTable(p storedProc.UserTokenSyncRecord) {
	if p.Rowid <= 0 {
		return
	}

	updateSQL := fmt.Sprintf("update %s.%s a set status='%s' , statusmessage = '%s' where rrn(a) = %d", s.UserTokenFileLib, s.UserTokenFile, p.Status, p.StatusMessage, p.Rowid)
	conn, err := s.GetSingleConnection()
	defer conn.Close()
	if err != nil {
		log.Println("Error updating User token file....", err.Error())
	}
	_, err = conn.Exec(updateSQL)
	if err != nil {
		log.Println("Error updateing User token file.... ", err.Error())
	}
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) SyncUserTokenRecords(withupdate bool) ([]*storedProc.UserTokenSyncRecord, error) {

	userTokens := make([]*storedProc.UserTokenSyncRecord, 0)
	if strings.TrimSpace(s.UserTokenFile) != "" && strings.TrimSpace(s.UserTokenFileLib) != "" {

		sqlToUse := fmt.Sprintf("select rrn(a), upper(trim(username)) , trim(token) from %s.%s a where status=''", s.UserTokenFileLib, s.UserTokenFile)

		conn, err := s.GetSingleConnection()
		defer conn.Close()
		if err != nil {

			return userTokens, err
		}

		rows, err := conn.Query(sqlToUse)

		defer rows.Close()
		if err != nil {
			// var odbcError *odbc.Error

			// if errors.As(err, &odbcError) {
			// 	s.UpdateAfterError(odbcError)
			// }
			return userTokens, err
		}

		for rows.Next() {
			rcd := &storedProc.UserTokenSyncRecord{}
			err := rows.Scan(&rcd.Rowid,
				&rcd.Username,
				&rcd.Token,
			)
			if err != nil {
				rcd.Status = "E"
				rcd.StatusMessage = err.Error()
			} else {
				rcd.CheckField(validator.NotBlank(rcd.Username), "ErrorMsg", "Username: This field cannot be blank")
				rcd.CheckField(validator.NotBlank(rcd.Token), "ErrorMsg", "Token: This field cannot be blank")

				if rcd.Valid() {
					rcd.Status = "P"
					rcd.StatusMessage = "processing"

				} else {
					// update table with error
					rcd.Status = "E"
					rcd.StatusMessage = rcd.Validator.FieldErrors["ErrorMsg"]
				}
			}
			userTokens = append(userTokens, rcd)
		}
	}
	// if withupdate && updateSQL != "" {
	// 	_, err := conn.Exec(updateSQL)
	// 	if err != nil {
	// 		log.Println("Error updateing promotion status ", err.Error())
	// 	}
	// }

	return userTokens, nil
}
