package ibmiServer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/onlysumitg/godbc"
	"github.com/zerobit-tech/GoQhttp/env"

	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	"github.com/zerobit-tech/GoQhttp/internal/validator"
	"github.com/zerobit-tech/GoQhttp/logger"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
	"github.com/zerobit-tech/GoQhttp/utils/xmlutils"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *Server) Refresh(ctx context.Context, sp *storedProc.StoredProc) error {
	if s.hasSPUpdated(ctx, sp) {
		err := s.PrepareToSave(ctx, sp)
		if err != nil {
			return err
		}
	}

	return nil
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
// func (s *Server) PromotionRecordToStoredProcX(p storedProc.PromotionRecord) *storedProc.StoredProc {
// 	sp := &storedProc.StoredProc{
// 		EndPointName: p.Endpoint,
// 		HttpMethod:   p.Httpmethod,
// 		Name:         p.Storedproc,
// 		Lib:          p.Storedproclib,
// 	}
// 	if p.UseSpecificName == "Y" {
// 		sp.UseSpecificName = true
// 	}

// 	if p.UseWithoutAuth == "Y" {
// 		sp.AllowWithoutAuth = true
// 	}
// 	srcd := &storedProc.ServerRecord{
// 		ID:   s.ID,
// 		Name: s.Name,
// 	}
// 	sp.DefaultServer = srcd

// 	return sp
// }

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *Server) PrepareToSave(ctx context.Context, sp *storedProc.StoredProc) error {
	sp.Name = strings.ToUpper(strings.TrimSpace(sp.Name))
	sp.Lib = strings.ToUpper(strings.TrimSpace(sp.Lib))
	sp.HttpMethod = strings.ToUpper(strings.TrimSpace(sp.HttpMethod))
	sp.UseNamedParams = true

	ctx1, cancelFunc1 := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc1()
	err := s.getResultSetCount(ctx1, sp)
	if err != nil {
		return err
	}

	ctx2, cancelFunc2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc2()
	err = s.getParameters(ctx2, sp)
	if err != nil {
		return err
	}

	for _, p := range sp.Parameters {
		if isUnsupportedDataType(p.Datatype, p.Mode) {
			return fmt.Errorf("%s %s (datatype %s) not supported", p.Mode, p.Name, p.Datatype)
		}
	}

	s.buildCallStatement(sp, sp.UseNamedParams)
	sp.BuildMockUrl()

	s.buildPromotionSQL(sp)

	return nil
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetConnectionString() string {
	// https://www.ibm.com/docs/en/i/7.4?topic=details-connection-string-keywords

	connectionString := ""

	if strings.HasPrefix(strings.ToUpper(s.IP), "*DSN:") {
		connectionString = s.getDNSConnectionString()
	} else {
		connectionString = s.getIPConnectionString()
	}

	return connectionString
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) getIPConnectionString() string {
	driver := "IBM i Access ODBC Driver"
	ssl := 0
	if s.Ssl {
		ssl = 1
	}
	pwd := s.GetPassword()

	libList := s.GetConnetionLibList()

	return fmt.Sprintf("DRIVER=%s;SYSTEM=%s;UID=%s;PWD=%s;DBQ=%s;UNICODESQL=1;XDYNAMIC=1;EXTCOLINFO=0;PKG=A/QHTTP,2,0,0,1,512;PROTOCOL=TCPIP;NAM=1;CMT=0;SSL=%d;ALLOWUNSCHAR=1", driver, s.IP, s.GetUserName(), pwd, libList, ssl)

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) getDNSConnectionString() string {
	dnsName := fmt.Sprintf("DSN=%s;", strings.TrimSpace(s.IP[5:]))

	userName := s.GetUserName()
	userString := ""
	if !strings.EqualFold(userName, "*DSN") {
		userString = fmt.Sprintf("UID=%s;", userName)
	}

	pwd := s.GetPassword()
	pwdString := ""
	if !strings.EqualFold(pwd, "*DSN") {
		pwdString = fmt.Sprintf("PWD=%s;", pwd)
	}

	libList := s.GetConnetionLibList()

	return fmt.Sprintf("%s%s%sDBQ=%s;UNICODESQL=1;XDYNAMIC=1;EXTCOLINFO=0;CMT=0;ALLOWUNSCHAR=1;NAM=1", dnsName, userString, pwdString, libList)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetPassword() string {
	pwd, err := stringutils.Decrypt(s.Password, s.GetSecretKey())
	if err != nil {
		log.Println("Unable to decrypt password")
		return ""
	}
	if strings.ToUpper(pwd) == "*ENV" {
		pwd = env.GetServerPassword(s.Name)

	}

	return pwd
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetConnectionType() string {
	return "godbc" //"odbc"
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) PingTimeoutDuration() time.Duration {
	age := 3
	if s.PingTimeout > 0 {
		age = s.PingTimeout
	}

	return time.Duration(age) * time.Second
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetConnectionID() string {
	return s.ID
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) GetSecretKey() string {
	return "Ang&1*~U^2^#s0^=)^^7#b34"
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *Server) APICall(ctx context.Context, callID string, sp *storedProc.StoredProc, params map[string]xmlutils.ValueDatatype, paramRegex map[string]string) (responseFormat *storedProc.StoredProcResponse, callDuration time.Duration, err error) {
	//log.Printf("%v: %v\n", "SeversCall005.001", time.Now())

	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	t1 := time.Now()
	defer func() {
		if r := recover(); r != nil {
			responseFormat = &storedProc.StoredProcResponse{
				ReferenceId: "string",
				Status:      500,
				Message:     fmt.Sprintf("%s", r),
				Data:        map[string]any{},
				//LogData:     []storedProc.LogByType{{Text: fmt.Sprintf("%s", r), Type: "ERROR"}},
			}
			responseFormat.LogData = []*logger.LogEvent{logger.GetLogEvent("ERROR", callID, fmt.Sprintf("%s", r), false)}
			callDuration = time.Since(t1)
			// apiCall.Response = responseFormat
			err = fmt.Errorf("%s", r)
		}
	}()

	givenParams := make(map[string]any)
	//.LogInfo("Building parameters for SP call")
	for k, v := range params {
		givenParams[k] = v.Value
	}
	return s.call(ctx, callID, sp, givenParams, paramRegex)

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *Server) DummyCall(sp *storedProc.StoredProc, givenParams map[string]any, paramRegex map[string]string) (*storedProc.StoredProcResponse, error) {
	preparedCallStatements, err := s.prepareCallStatement(sp, givenParams, paramRegex)
	if err != nil {
		return nil, err
	}
	err = s.SeversCall(context.Background(), sp, preparedCallStatements, true)
	if err != nil {
		return nil, err
	}

	responseFormat := &storedProc.StoredProcResponse{
		ReferenceId: "string",
		Status:      200,
		Message:     "string",
		Data:        preparedCallStatements.ResponseFormat,
	}

	b, err := json.MarshalIndent(responseFormat, "", "\t")

	if err != nil {
		return nil, err

	}
	sp.ResponseFormat = string(b)
	return responseFormat, nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *Server) Exists(ctx context.Context, sp *storedProc.StoredProc) (bool, error) {

	exists := "N"

	sqlToRun := fmt.Sprintf("select 'Y'  from qsys2.sysprocs where SPECIFIC_NAME='%s'  and SPECIFIC_SCHEMA='%s'    limit 1", strings.ToUpper(sp.SpecificName), strings.ToUpper(sp.SpecificLib))

	//sqlToRun = fmt.Sprintf("select Y  from qsys2.sysprocs where ROUTINE_NAME='%s'  and ROUTINE_SCHEMA='%s'   limit 1", strings.ToUpper(sp.Name), strings.ToUpper(sp.Lib))

	conn, err := s.GetConnection()

	if err != nil {
		return true, err // to prevent delete
	}
	row := conn.QueryRowContext(ctx, sqlToRun)

	err = row.Scan(&exists)

	if err != nil {
		return true, err

	}

	if exists == "Y" {
		return true, nil
	}

	return false, nil
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) ErrorToHttpStatus(inerr error) (int, string, string, bool) {
	var odbcError *godbc.Error

	if errors.As(inerr, &odbcError) {

		if len(odbcError.Diag) > 0 {
			code := odbcError.Diag[0].NativeError
			switch code {
			case -420:
				return http.StatusBadRequest, "Please check the values.", odbcError.Error(), true
			case -204:
				return http.StatusNotFound, "OD0204[42S02]", odbcError.Error(), true
			case 8001:
				return http.StatusInternalServerError, "OD8001", odbcError.Error(), true
			case 10060:
				return http.StatusInternalServerError, "OD10060", odbcError.Error(), true
			case 30038:
				return http.StatusInternalServerError, "OD30038", odbcError.Error(), true
			case 30189:
				return http.StatusInternalServerError, "OD30189", odbcError.Error(), true // {HYT00} [IBM][System i Access ODBC Driver]Connection login timed out.
			case 10065:
				return http.StatusInternalServerError, "OD10065", odbcError.Error(), true // "[IBM][System i Access ODBC Driver]Communication link failure. comm rc=10065 - CWBCO1003 - Sockets error, function  returned 10065 "
			case 8002:
				return http.StatusInternalServerError, "OD8002", odbcError.Error(), true // SQLDriverConnect: {28000} [IBM][System i Access ODBC Driver]Communication link failure. comm rc=8002 - CWBSY0002 - Password for user SUMITG33 on system PUB400.COM is not correct, Password length = 10, Prompt Mode = Never, System IP Address = 185.113.5.134

			default:
				return http.StatusInternalServerError, fmt.Sprintf("OD%d", code), odbcError.Error(), true // SQLDriverConnect: {28000} [IBM][System i Access ODBC Driver]Communication link failure. comm rc=8002 - CWBSY0002 - Password for user SUMITG33 on system PUB400.COM is not correct, Password length = 10, Prompt Mode = Never, System IP Address = 185.113.5.134
			}

		}

		return http.StatusBadRequest, odbcError.Error(), odbcError.Error(), true
	}
	return 0, "", "", false

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) SyncUserTokenRecords(withupdate bool) ([]*storedProc.UserTokenSyncRecord, error) {

	userTokens := make([]*storedProc.UserTokenSyncRecord, 0)
	if strings.TrimSpace(s.UserTokenFile) != "" && strings.TrimSpace(s.UserTokenFileLib) != "" {

		sqlToUse := fmt.Sprintf("select rrn(a), upper(trim(ifnull(useremail,''))) , trim(ifnull(token,'')) from %s.%s a where ifnull(status,'')=''", s.UserTokenFileLib, s.UserTokenFile)

		conn, err := s.GetSingleConnection()
		if err != nil {

			return userTokens, err
		}
		defer conn.Close()

		rows, err := conn.Query(sqlToUse)

		if err != nil {
			// var odbcError *odbc.Error

			// if errors.As(err, &odbcError) {
			// 	s.UpdateAfterError(odbcError)
			// }
			return userTokens, err
		}
		defer rows.Close()

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

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) UpdateStatusForPromotionRecord(p storedProc.PromotionRecord) {
	if p.Rowid == "0" || p.Rowid == "" {
		return
	}

	updateSQL := fmt.Sprintf("update %s.%s a set status='%s' , statusmessage = '%s' where rrn(a) = %s", s.ConfigFileLib, s.ConfigFile, p.Status, p.StatusMessage, p.Rowid)
	conn, err := s.GetSingleConnection()
	if err != nil {
		log.Println("Error updating promotion file....", err.Error())
	}
	defer conn.Close()
	_, err = conn.Exec(updateSQL)
	if err != nil {
		log.Println("Error updateing promotion file.... ", err.Error())
	}
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) UpdateStatusUserTokenTable(p storedProc.UserTokenSyncRecord) {
	if p.Rowid == "0" || p.Rowid == "" {
		return
	}

	updateSQL := fmt.Sprintf("update %s.%s a set status='%s' , statusmessage = '%s' where rrn(a) = %s", s.UserTokenFileLib, s.UserTokenFile, p.Status, p.StatusMessage, p.Rowid)
	conn, err := s.GetSingleConnection()
	if err != nil {
		log.Println("Error updating User token file....", err.Error())
	}
	defer conn.Close()
	_, err = conn.Exec(updateSQL)
	if err != nil {
		log.Println("Error updateing User token file.... ", err.Error())
	}
}
