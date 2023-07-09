package ibmiServer

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/onlysumitg/GoQhttp/dbserver"
	"github.com/onlysumitg/GoQhttp/go_ibm_db"
	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/logger"
	"github.com/onlysumitg/GoQhttp/utils/httputils"
	"github.com/onlysumitg/GoQhttp/utils/stringutils"
	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
)

type IBMiServer struct {
	*dbserver.Server
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *IBMiServer) Load(bs *dbserver.Server) {
	s.Server = bs

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *IBMiServer) Refresh(ctx context.Context, sp *storedProc.StoredProc) error {
	if s.HasSPUpdated(ctx, sp) {
		err := s.PreapreToSave(ctx, sp)
		if err != nil {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *IBMiServer) HasSPUpdated(ctx context.Context, sp *storedProc.StoredProc) bool {

	hasModified := "N"

	sqlToRun := fmt.Sprintf("select 'Y'  from qsys2.sysprocs where SPECIFIC_NAME='%s'  and SPECIFIC_SCHEMA='%s'  and ROUTINE_CREATED != '%s' limit 1", strings.ToUpper(sp.SpecificName), strings.ToUpper(sp.SpecificLib), sp.Modified)
	// } else {
	// 	sqlToRun = fmt.Sprintf("select Y  from qsys2.sysprocs where ROUTINE_NAME='%s'  and ROUTINE_SCHEMA='%s'  and ROUTINE_CREATED != '%s' limit 1", strings.ToUpper(sp.Name), strings.ToUpper(sp.Lib), sp.Modified)

	// }

	conn, err := s.GetConnection()

	if err != nil {
		return false
	}
	row := conn.QueryRowContext(ctx, sqlToRun)

	err = row.Scan(&hasModified)

	if err != nil {
		return false

	}

	if hasModified == "Y" {
		return true
	}

	return false
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *IBMiServer) Exists(ctx context.Context, sp *storedProc.StoredProc) (bool, error) {

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

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *IBMiServer) PreapreToSave(ctx context.Context, sp *storedProc.StoredProc) error {
	sp.Name = strings.ToUpper(strings.TrimSpace(sp.Name))
	sp.Lib = strings.ToUpper(strings.TrimSpace(sp.Lib))
	sp.HttpMethod = strings.ToUpper(strings.TrimSpace(sp.HttpMethod))
	sp.UseNamedParams = true

	ctx1, cancelFunc1 := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc1()
	err := s.GetResultSetCount(ctx1, sp)
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

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *IBMiServer) buildCallStatement(sp *storedProc.StoredProc, useNamedParams bool) (err error) {

	paramString := ""

	for _, parameter := range sp.Parameters {
		value := ""
		switch parameter.Mode {
		case "IN":
			value = fmt.Sprintf("'{:%s}'", parameter.Name) //fmt.Sprintf("'%s'", parameter.GivenValue)
		case "OUT":
			value = "?"
		case "INOUT":
			value = "?"
		}

		if useNamedParams {
			paramString += fmt.Sprintf("%s=>%s %s", parameter.Name, value, ",")
		} else {
			paramString += fmt.Sprintf("%s ,", value)
		}

	}

	paramString = strings.TrimRight(paramString, ",")
	sp.CallStatement = fmt.Sprintf("call %s.%s(%s)", sp.SpecificLib, sp.SpecificName, paramString)
	return nil

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *IBMiServer) prepareCallStatement(sp *storedProc.StoredProc, givenParams map[string]any) (*storedProc.PreparedCallStatements, error) {

	spResponseFormat := make(map[string]any)
	inoutParams := make([]any, 0)
	inoutParamVariables := make(map[string]*any)
	inOutParamMapToSPParam := make(map[string]*storedProc.StoredProcParamter)

	finalCallStatement := sp.CallStatement

	for _, p := range sp.Parameters {
		paramNameToUse := p.GetNameToUse()

		switch p.Mode {
		case "IN":
			valueToUse, found := givenParams[paramNameToUse]
			if !found {
				valueToUse = s.GetDefaultValue(p)

			} else {
				//p.GivenValue = asString(valueToUse)

			}
			if !p.HasValidValue(valueToUse) {
				return nil, fmt.Errorf("%s: invalid value", p.Name)
			}

			stringToReplace := ""
			if p.NeedQuote(stringutils.AsString(valueToUse)) {
				stringToReplace = fmt.Sprintf("{:%s}", p.Name)
			} else {
				stringToReplace = fmt.Sprintf("'{:%s}'", p.Name)
			}
			stringToReplace = fmt.Sprintf("'{:%s}'", p.Name)
			inoutParams = append(inoutParams, &valueToUse)

			finalCallStatement = strings.ReplaceAll(finalCallStatement, stringToReplace, "?")

		case "INOUT":
			spResponseFormat[p.GetNameToUse()] = p.Datatype

			valueToUse, found := givenParams[paramNameToUse]
			if !found {
				valueToUse = s.GetDefaultValue(p)
				if valueToUse == "NULL" {
					valueToUse = nil
				}

			} else {
				//p.GivenValue = asString(valueToUse)

			}
			if !p.HasValidValue(valueToUse) {
				return nil, fmt.Errorf("%s: invalid value", stringutils.AsString(valueToUse))
			}

			valueToUse, err := p.ConvertToType(valueToUse)
			if err != nil {
				return nil, fmt.Errorf("%s: invalid value", stringutils.AsString(valueToUse))

			}
			inoutParamVariables[p.Name] = &valueToUse

			inoutParams = append(inoutParams, sql.Out{Dest: inoutParamVariables[p.Name], In: true})

			inOutParamMapToSPParam[p.Name] = p

		case "OUT":
			spResponseFormat[p.GetNameToUse()] = p.Datatype
			out := p.GetofType()
			inoutParamVariables[p.Name] = out
			inoutParams = append(inoutParams, sql.Out{Dest: inoutParamVariables[p.Name]})
			inOutParamMapToSPParam[p.Name] = p

		}
	}

	return &storedProc.PreparedCallStatements{
		ResponseFormat:         spResponseFormat,
		InOutParams:            inoutParams,
		InOutParamVariables:    inoutParamVariables,
		InOutParamMapToSPParam: inOutParamMapToSPParam,
		FinalCallStatement:     finalCallStatement,
	}, nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *IBMiServer) APICall(ctx context.Context, callID string, sp *storedProc.StoredProc, params map[string]xmlutils.ValueDatatype) (responseFormat *storedProc.StoredProcResponse, callDuration time.Duration, err error) {
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
func (s *IBMiServer) Call(ctx context.Context, callID string, sp *storedProc.StoredProc, givenParams map[string]any) (*storedProc.StoredProcResponse, time.Duration, error) {
	//log.Printf("%v: %v\n", "SeversCall005.002", time.Now())
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	qhttp_status_code := 200
	qhttp_status_message := ""

	logEntries := make([]*logger.LogEvent, 0, 5)
	preparedCallStatements, err := s.prepareCallStatement(sp, givenParams)
	if err != nil {
		logEntries = append(logEntries, logger.GetLogEvent("ERROR", callID, err.Error(), false))
		return &storedProc.StoredProcResponse{LogData: logEntries}, 0, err
	}

	t1 := time.Now()
	logEntries = append(logEntries, logger.GetLogEvent("INFO", callID, "Starting DB CALL", false))

	err = s.SeversCall(ctx, sp, preparedCallStatements, false)

	spCallDuration := time.Since(t1)

	logEntries = append(logEntries, logger.GetLogEvent("INFO", callID, fmt.Sprintf("Finished DB CALL in: %s", spCallDuration), false))

	//log.Printf("%v: %v\n", "SeversCall005.004", time.Now())

	if err != nil {
		logEntries = append(logEntries, logger.GetLogEvent("ERROR", callID, err.Error(), false))

		return &storedProc.StoredProcResponse{LogData: logEntries}, 0, err
	}
	logEntries = append(logEntries, logger.GetLogEvent("INFO", callID, "SP Call complete", false))

	// read INOUT and OUT parameter values
	for kXX, v := range preparedCallStatements.InOutParamVariables {
		p, found := preparedCallStatements.InOutParamMapToSPParam[kXX]

		keyToUse := p.GetNameToUse()

		if found {

			if p.IsString() || reflect.ValueOf(v).Kind() == reflect.String {

				b, ok := (*v).([]byte)
				if ok {
					strVal := string(b)
					assignStrVal := true
					if validator.MustBeJSON(strVal) {
						jsonData := make(map[string]any)
						err := json.Unmarshal(b, &jsonData)

						if err == nil {
							preparedCallStatements.ResponseFormat[keyToUse] = &jsonData
							assignStrVal = false
						}

					}
					if assignStrVal {
						preparedCallStatements.ResponseFormat[keyToUse] = strVal

					}

					if p.Mode == "OUT" && keyToUse == "QHTTP_STATUS_MESSAGE" {
						qhttp_status_message = strVal
						delete(preparedCallStatements.ResponseFormat, keyToUse)

					}

				} else {
					preparedCallStatements.ResponseFormat[keyToUse] = v
				}
			} else {
				cv, err := p.ConvertOUTVarToType(v)
				if err == nil {
					preparedCallStatements.ResponseFormat[keyToUse] = cv
				} else {
					preparedCallStatements.ResponseFormat[keyToUse] = v
				}

			}

			if p.Mode == "OUT" && keyToUse == "QHTTP_STATUS_CODE" && p.IsInt() {
				intval, ok := 0, false

				switch reflect.ValueOf(*v).Kind() {
				case reflect.Int32:
					if intval32, ok2 := (*v).(int32); ok2 {
						intval = int(intval32)
						ok = ok2
					}
				case reflect.Int64:
					if intval64, ok2 := (*v).(int64); ok2 {
						intval = int(intval64)
						ok = ok2
					}
				case reflect.Int16:
					if intval16, ok2 := (*v).(int16); ok2 {
						intval = int(intval16)
						ok = ok2
					}
				case reflect.Int8:
					if intval8, ok2 := (*v).(int8); ok2 {
						intval = int(intval8)
						ok = ok2
					}
				default:
					intval, ok = (*v).(int)
				}

				if ok {
					validCode, message := httputils.IsValidHttpCode(int(intval))
					if validCode {
						qhttp_status_code = int(intval)

						// remove QHTTP_STATUS_CODE from out params
						delete(preparedCallStatements.ResponseFormat, keyToUse)

						if qhttp_status_message == "" {
							qhttp_status_message = message
						}
					}
				}

			}

		}
	}

	responseFormat := &storedProc.StoredProcResponse{
		ReferenceId: "string",
		Status:      qhttp_status_code,
		Message:     qhttp_status_message,
		Data:        preparedCallStatements.ResponseFormat,
		LogData:     logEntries,
	}

	return responseFormat, spCallDuration, nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *IBMiServer) DummyCall(sp *storedProc.StoredProc, givenParams map[string]any) (*storedProc.StoredProcResponse, error) {
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
func (s *IBMiServer) SeversCall(ctx context.Context, sp *storedProc.StoredProc, preparedCallStatements *storedProc.PreparedCallStatements, dummyCall bool) (ferr error) {

	defer func() {
		if r := recover(); r != nil {
			ferr = fmt.Errorf("%s", r)
		}
	}()

	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	db, err := s.GetConnection()
	if err != nil {
		return err
	}

	resultsets := make(map[string][]map[string]any, 0)
	ctx = context.WithValue(ctx, go_ibm_db.LOAD_SP_RESULT_SETS, resultsets)
	ctx = context.WithValue(ctx, go_ibm_db.DUMMY_SP_CALL, dummyCall)
	//ctx = context.WithValue(ctx, go_ibm_db.ESCAPE_QUOTE, true)  // use strconv.Quote on result set

	_, err = db.ExecContext(ctx, preparedCallStatements.FinalCallStatement, preparedCallStatements.InOutParams...)

	if err != nil {
		return err
	}

	// assign result sets

	for k, v := range resultsets {
		preparedCallStatements.ResponseFormat[k] = v

	}
	//	preparedCallStatements.ResponseFormat["data"] = resultsets

	return nil

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *IBMiServer) GetResultSetCount(ctx context.Context, sp *storedProc.StoredProc) error {

	resultSets := 0

	sp.ResultSets = 0

	sqlToRun := ""
	if sp.UseSpecificName {
		sqlToRun = fmt.Sprintf("select trim(SPECIFIC_SCHEMA), trim(SPECIFIC_NAME),trim(ROUTINE_SCHEMA),trim(ROUTINE_NAME), RESULT_SETS, SQL_DATA_ACCESS,ROUTINE_CREATED from qsys2.sysprocs where SPECIFIC_NAME='%s'  and SPECIFIC_SCHEMA='%s' limit 1", strings.ToUpper(sp.Name), strings.ToUpper(sp.Lib))
	} else {
		sqlToRun = fmt.Sprintf("select trim(SPECIFIC_SCHEMA), trim(SPECIFIC_NAME),trim(ROUTINE_SCHEMA),trim(ROUTINE_NAME), RESULT_SETS, SQL_DATA_ACCESS,ROUTINE_CREATED from qsys2.sysprocs where ROUTINE_NAME='%s'  and ROUTINE_SCHEMA='%s' limit 1", strings.ToUpper(sp.Name), strings.ToUpper(sp.Lib))

	}

	conn, err := s.GetConnection()

	if err != nil {
		return err
	}
	row := conn.QueryRowContext(ctx, sqlToRun)

	err = row.Scan(&sp.SpecificLib, &sp.SpecificName, &sp.Lib, &sp.Name, &resultSets, &sp.DataAccess, &sp.Modified)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Not found")
		}
		return err

	}
	sp.ResultSets = resultSets

	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *IBMiServer) GetParameters(ctx context.Context, sp *storedProc.StoredProc) error {

	originalParams := sp.Parameters

	sqlToUse := fmt.Sprintf("SELECT ORDINAL_POSITION, upper(trim(PARAMETER_MODE)) , upper(trim(PARAMETER_NAME)),DATA_TYPE, ifnull(NUMERIC_SCALE,0), ifnull(NUMERIC_PRECISION,0), ifnull(CHARACTER_MAXIMUM_LENGTH,0),  default FROM qsys2.sysparms WHERE SPECIFIC_NAME='%s' and   SPECIFIC_SCHEMA ='%s' ORDER BY ORDINAL_POSITION", strings.ToUpper(sp.SpecificName), strings.ToUpper(sp.SpecificLib))

	sp.Parameters = make([]*storedProc.StoredProcParamter, 0)
	conn, err := s.GetConnection()

	if err != nil {

		return err
	}

	rows, err := conn.QueryContext(ctx, sqlToUse)

	defer func() {
		rows.Close()
	}()

	if err != nil {
		// var odbcError *odbc.Error

		// if errors.As(err, &odbcError) {
		// 	s.UpdateAfterError(odbcError)
		// }
		return err
	}

	for rows.Next() {
		spParamter := &storedProc.StoredProcParamter{}
		err := rows.Scan(&spParamter.Position,
			&spParamter.Mode,
			&spParamter.Name,
			&spParamter.Datatype,
			&spParamter.Scale,
			&spParamter.Precision,
			&spParamter.MaxLength,
			&spParamter.DefaultValue)
		if err != nil {
			//log.Println("GetSPParameter ", err.Error())
		}

		if strings.TrimSpace(spParamter.Datatype) == "" {
			spParamter.Datatype = "CHARACTER"
		}

		if strings.TrimSpace(spParamter.Name) == "" {
			sp.UseNamedParams = false
			spParamter.Name = strconv.Itoa(spParamter.Position)

		}
		sp.Parameters = append(sp.Parameters, spParamter)

	}

	// restore alias
	for _, p := range sp.Parameters {
		for _, op := range originalParams {
			if p.Name == op.Name {
				p.Alias = op.Alias
			}
		}
	}

	return nil

}
