package models

import (
	"fmt"
	"log"
	"strings"

	"github.com/onlysumitg/GoQhttp/internal/validator"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

type UserTokenSyncRecord struct {
	Rowid               int
	Username            string // D: Delete   R:Refresh   I:Insert
	Token               string
	Status              string
	StatusMessage       string
	validator.Validator `json:"-" db:"-" form:"-"`
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p UserTokenSyncRecord) UpdateStatusUserTokenTable(s *Server) {
	if p.Rowid <= 0 {
		return
	}

	updateSQL := fmt.Sprintf("update %s.%s a set status='%s' , statusmessage = '%s' where rrn(a) = %d", s.UserTokenFileLib, s.UserTokenFile, p.Status, p.StatusMessage, p.Rowid)
	conn, err := s.GetConnection()

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
func (s *Server) SyncUserTokenRecords(withupdate bool) ([]*UserTokenSyncRecord, error) {

	userTokens := make([]*UserTokenSyncRecord, 0)
	if strings.TrimSpace(s.UserTokenFile) != "" && strings.TrimSpace(s.UserTokenFileLib) != "" {

		sqlToUse := fmt.Sprintf("select rrn(a), upper(trim(username)) , trim(token) from %s.%s a where status=''", s.UserTokenFileLib, s.UserTokenFile)

		conn, err := s.GetConnection()

		if err != nil {

			return userTokens, err
		}

		rows, err := conn.Query(sqlToUse)
		if err != nil {
			// var odbcError *odbc.Error

			// if errors.As(err, &odbcError) {
			// 	s.UpdateAfterError(odbcError)
			// }
			return userTokens, err
		}

		for rows.Next() {
			rcd := &UserTokenSyncRecord{}
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
