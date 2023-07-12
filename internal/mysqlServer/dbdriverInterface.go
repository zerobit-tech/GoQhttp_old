package mysqlserver

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

	"github.com/onlysumitg/GoQhttp/go_ibm_db"
	"github.com/onlysumitg/GoQhttp/internal/dbserver"
	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/logger"
	"github.com/onlysumitg/GoQhttp/utils/stringutils"
	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MySqlServer) LoadX(bs *dbserver.Server) {
	s.Server = bs

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MySqlServer) RefreshX(ctx context.Context, sp *storedProc.StoredProc) error {
	if s.HasSPUpdated(ctx, sp) {
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
// func (s *MySqlServer) PromotionRecordToStoredProcX(p storedProc.PromotionRecord) *storedProc.StoredProc {
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
func (s *MySqlServer) PrepareToSaveX(ctx context.Context, sp *storedProc.StoredProc) error {
	sp.Name = strings.ToUpper(strings.TrimSpace(sp.Name))
	sp.Lib = strings.ToUpper(strings.TrimSpace(sp.Lib))
	sp.HttpMethod = strings.ToUpper(strings.TrimSpace(sp.HttpMethod))
	sp.UseNamedParams = true

	ctx1, cancelFunc1 := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc1()
	err := s.getSPDetails(ctx1, sp)
	if err != nil {
		return err
	}

	ctx2, cancelFunc2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc2()
	err = s.GetParameters(ctx2, sp)
	if err != nil {
		return err
	}

	for _, p := range sp.Parameters {
		if IsUnsupportedDataType(p.Datatype, p.Mode) {
			return fmt.Errorf("%s %s (datatype %s) not supported", p.Mode, p.Name, p.Datatype)
		}
	}

	s.buildCallStatement(sp, sp.UseNamedParams)
	sp.BuildMockUrl()

	s.BuildPromotionSQL(sp)

	return nil
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MySqlServer) GetConnectionStringX() string {

	pwd := s.GetPassword()
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/?multiStatements=true&autocommit=true", s.UserName, pwd, s.IP, s.Port)
	//connectionString := fmt.Sprintf("DSN=pub400; UID=%s;PWD=%s", s.UserName, s.Password)

	return connectionString
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MySqlServer) GetPasswordX() string {
	pwd, err := stringutils.Decrypt(s.Password, s.GetSecretKey())
	if err != nil {
		log.Println("Unable to decrypt password")
		return ""
	}
	return pwd
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MySqlServer) GetConnectionTypeX() string {
	return "mysql" //"odbc"
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MySqlServer) PingTimeoutDurationX() time.Duration {
	age := 3
	if s.PingTimeout > 0 {
		age = s.PingTimeout
	}

	return time.Duration(age) * time.Second
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MySqlServer) GetConnectionIDX() string {
	return s.ID
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MySqlServer) GetSecretKeyX() string {
	return "Aql&1*~P^2^#y0^=)^^7#x34"
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MySqlServer) APICallX(ctx context.Context, callID string, sp *storedProc.StoredProc, params map[string]xmlutils.ValueDatatype) (responseFormat *storedProc.StoredProcResponse, callDuration time.Duration, err error) {
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
	return s.Call(ctx, callID, sp, givenParams)

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MySqlServer) DummyCallX(sp *storedProc.StoredProc, givenParams map[string]any) (*storedProc.StoredProcResponse, error) {
	preparedCallStatements, err := s.prepareCallStatement(sp, givenParams)
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
func (s *MySqlServer) ExistsX(ctx context.Context, sp *storedProc.StoredProc) (bool, error) {

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
func (s *MySqlServer) ErrorToHttpStatusX(inerr error) (int, string, string, bool) {
	var odbcError *go_ibm_db.Error

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
			}

		}

		return http.StatusBadRequest, odbcError.Error(), odbcError.Error(), true
	}
	return 0, "", "", false

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MySqlServer) ListPromotionX(withupdate bool) ([]*storedProc.PromotionRecord, error) {

	promotionRecords := make([]*storedProc.PromotionRecord, 0)
	if strings.TrimSpace(s.ConfigFile) != "" && strings.TrimSpace(s.ConfigFileLib) != "" {

		sqlToUse := fmt.Sprintf("select rrn(a), upper(trim(ifnull(operation,''))) , upper(trim(ifnull(endpoint,''))), trim(ifnull(storedproc,'')), trim(ifnull(storedproclib,'')), upper(trim(ifnull(httpmethod,''))), upper(trim(ifnull(usespecificname,''))), upper(trim(ifnull(usewithoutauth,''))) , upper(trim(ifnull(paramalias,''))) from %s.%s a where ifnull(status,'')=''", s.ConfigFileLib, s.ConfigFile)

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
func (s *MySqlServer) SyncUserTokenRecordsX(withupdate bool) ([]*storedProc.UserTokenSyncRecord, error) {

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
func (s *MySqlServer) UpdateStatusForPromotionRecordX(p storedProc.PromotionRecord) {
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
func (s *MySqlServer) UpdateStatusUserTokenTableX(p storedProc.UserTokenSyncRecord) {
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
