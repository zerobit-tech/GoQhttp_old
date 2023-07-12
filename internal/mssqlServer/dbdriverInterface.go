package mssqlserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/onlysumitg/GoQhttp/internal/dbserver"
	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/utils/stringutils"
	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MSSqlServer) LoadX(bs *dbserver.Server) {
	s.Server = bs

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) GetConnectionStringX() string {

	pwd := s.GetPassword()

	//connectionString := fmt.Sprintf("DSN=pub400; UID=%s;PWD=%s", s.UserName, s.Password)
	connectionString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d", s.IP, s.UserName, pwd, s.Port)
	return connectionString
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) GetPasswordX() string {
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
func (s *MSSqlServer) GetConnectionTypeX() string {
	return "sqlserver"
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) PingTimeoutDurationX() time.Duration {
	age := 3
	if s.PingTimeout > 0 {
		age = s.PingTimeout
	}

	return time.Duration(age) * time.Second
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) GetConnectionIDX() string {
	return s.ID
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) GetSecretKeyX() string {
	return "BhL&1*~U^2^#s0^=)^^8#b34" // keep the length
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MSSqlServer) APICallX(ctx context.Context, callId string, sp *storedProc.StoredProc, params map[string]xmlutils.ValueDatatype) (responseFormat *storedProc.StoredProcResponse, callDuration time.Duration, err error) {
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
	return s.call(ctx, sp, givenParams)

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MSSqlServer) PrepareToSaveX(ctx context.Context, sp *storedProc.StoredProc) error {
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

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MSSqlServer) RefreshX(ctx context.Context, sp *storedProc.StoredProc) error {
	if s.hasSPUpdated(ctx, sp) {
		err := s.PrepareToSave(ctx, sp)
		if err != nil {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MSSqlServer) DummyCallX(sp *storedProc.StoredProc, givenParams map[string]any) (*storedProc.StoredProcResponse, error) {
	preparedCallStatements, err := s.prepareCallStatement(sp, givenParams)
	if err != nil {
		return nil, err
	}
	err = s.seversCall(context.Background(), sp, preparedCallStatements, true)
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

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) ErrorToHttpStatusX(inerr error) (httpcode int, returnToCustomer string, logText string, handledErro bool) {

	//return http.StatusBadRequest, odbcError.Error(), odbcError.Error(), true
	if strings.HasSuffix(inerr.Error(), "MSSQL does not allow NULL value without type for OUTPUT parameters") {
		return http.StatusBadRequest, "MS0001: NULL values not allowed", inerr.Error(), true
	}

	if strings.HasPrefix(inerr.Error(), "mssql: login error: Login failed for user") {
		return http.StatusInternalServerError, "MS0002: Server error", inerr.Error(), true
	}

	return 0, "", "", false

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MSSqlServer) ExistsX(ctx context.Context, sp *storedProc.StoredProc) (bool, error) {

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
func (s *MSSqlServer) UpdateStatusForPromotionRecordX(p storedProc.PromotionRecord) {
	if p.Rowid == "0" || p.Rowid == "" {
		return
	}

	updateCol := "sys.fn_PhysLocFormatter(%%physloc%%)"
	catalog, schma := breakStringToCatalogSchema(s.ConfigFileLib)

	updateSQL := fmt.Sprintf("update %s.%s.%s set status='%s' , statusmessage = '%s' where  %s = '%s'", catalog, schma, s.ConfigFile, p.Status, p.StatusMessage, updateCol, p.Rowid)

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
func (s *MSSqlServer) ListPromotionX(withupdate bool) ([]*storedProc.PromotionRecord, error) {

	promotionRecords := make([]*storedProc.PromotionRecord, 0)
	if strings.TrimSpace(s.ConfigFile) != "" && strings.TrimSpace(s.ConfigFileLib) != "" {

		catalog, schma := breakStringToCatalogSchema(s.ConfigFileLib)

		columns := "select sys.fn_PhysLocFormatter(%%physloc%%), upper(trim(isnull(operation,''))) , upper(trim(isnull(endpoint,''))), trim(isnull(storedproc,'')), trim(isnull(storedproclib,'')), upper(trim(isnull(httpmethod,''))), upper(trim(isnull(usespecificname,''))), upper(trim(isnull(usewithoutauth,''))) , upper(trim(isnull(paramalias,''))) "
		sqlToUse := fmt.Sprintf("%s from %s.%s.%s  where isnull(status,'')=''", columns, catalog, schma, s.ConfigFile)

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

	autoP, err := s.listAutoPromotion()
	if err == nil && len(autoP) > 0 {
		promotionRecords = append(promotionRecords, autoP...)
	}
	return promotionRecords, nil
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) UpdateStatusUserTokenTableX(p storedProc.UserTokenSyncRecord) {
	if p.Rowid == "0" || p.Rowid == "" {
		return
	}

	updateCol := "sys.fn_PhysLocFormatter(%%physloc%%)"
	catalog, schma := breakStringToCatalogSchema(s.UserTokenFileLib)

	updateSQL := fmt.Sprintf("update %s.%s.%s set status='%s' , statusmessage = '%s' where  %s = '%s'", catalog, schma, s.UserTokenFile, p.Status, p.StatusMessage, updateCol, p.Rowid)

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

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *MSSqlServer) SyncUserTokenRecordsX(withupdate bool) ([]*storedProc.UserTokenSyncRecord, error) {

	userTokens := make([]*storedProc.UserTokenSyncRecord, 0)
	if strings.TrimSpace(s.UserTokenFile) != "" && strings.TrimSpace(s.UserTokenFileLib) != "" {

		catalog, schma := breakStringToCatalogSchema(s.UserTokenFileLib)

		columns := "select sys.fn_PhysLocFormatter(%%physloc%%), upper(trim(isnull(useremail,''))) , upper(trim(isnull(token,''))) "
		sqlToUse := fmt.Sprintf("%s from %s.%s.%s  where isnull(status,'')=''", columns, catalog, schma, s.UserTokenFile)

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
