package mysqlserver

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

	"github.com/onlysumitg/GoQhttp/internal/dbserver"
	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/logger"
	"github.com/onlysumitg/GoQhttp/utils/httputils"
	"github.com/onlysumitg/GoQhttp/utils/stringutils"
)

type MySqlServer struct {
	*dbserver.Server
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MySqlServer) HasSPUpdated(ctx context.Context, sp *storedProc.StoredProc) bool {

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
// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MySqlServer) buildCallStatement(sp *storedProc.StoredProc, useNamedParams bool) (err error) {

	paramString := ""

	for _, parameter := range sp.Parameters {
		value := ""
		switch parameter.Mode {
		case "IN":
			value = fmt.Sprintf("'{:%s}'", parameter.Name) //fmt.Sprintf("'%s'", parameter.GivenValue)
		case "OUT":
			value = fmt.Sprintf("@%s", parameter.GetNameToUse())
		case "INOUT":
			value = fmt.Sprintf("@%s", parameter.GetNameToUse())
		}

		if useNamedParams {
			paramString += fmt.Sprintf("%s %s", value, ",")
		} else {
			paramString += fmt.Sprintf("%s ,", value)
		}

	}

	paramString = strings.TrimRight(paramString, ",")
	sp.CallStatement = fmt.Sprintf("call %s.%s(%s);", sp.SpecificLib, sp.SpecificName, paramString)
	return nil

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MySqlServer) prepareCallStatement(sp *storedProc.StoredProc, givenParams map[string]any) (*storedProc.PreparedCallStatements, error) {

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
			if !parameterHasValidValue(p, valueToUse) {
				return nil, fmt.Errorf("%s: invalid value", p.Name)
			}

			stringToReplace := ""
			if parameterNeedQuote(p, stringutils.AsString(valueToUse)) {
				stringToReplace = fmt.Sprintf("{:%s}", p.Name)
			} else {
				stringToReplace = fmt.Sprintf("'{:%s}'", p.Name)
			}
			//stringToReplace = fmt.Sprintf("'{:%s}'", p.Name)
			//inoutParams = append(inoutParams, &valueToUse)

			finalCallStatement = strings.ReplaceAll(finalCallStatement, stringToReplace, stringutils.AsString(valueToUse))

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
			if !parameterHasValidValue(p, valueToUse) {
				return nil, fmt.Errorf("%s: invalid value", stringutils.AsString(valueToUse))
			}

			valueToUse, err := p.ConvertToType(valueToUse)
			if err != nil {
				return nil, fmt.Errorf("%s: invalid value", stringutils.AsString(valueToUse))

			}
			inoutParamVariables[p.Name] = &valueToUse

			//	inoutParams = append(inoutParams, sql.Out{Dest: inoutParamVariables[p.Name], In: true})

			inOutParamMapToSPParam[p.Name] = p

		case "OUT":
			spResponseFormat[p.GetNameToUse()] = p.Datatype
			out := getParameterofType(p)
			inoutParamVariables[p.Name] = out
			//inoutParams = append(inoutParams, fmt.Sprintf("@%s", p.GetNameToUse()))
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
func (s *MySqlServer) Call(ctx context.Context, callID string, sp *storedProc.StoredProc, givenParams map[string]any) (*storedProc.StoredProcResponse, time.Duration, error) {
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
func (s *MySqlServer) SeversCall(ctx context.Context, sp *storedProc.StoredProc, preparedCallStatements *storedProc.PreparedCallStatements, dummyCall bool) (ferr error) {

	defer func() {
		if r := recover(); r != nil {
			ferr = fmt.Errorf("%s", r)
		}
	}()
	qhttp_status_code := 200
	qhttp_status_message := ""

	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	dbX, err := s.GetConnection()
	if err != nil {
		return err
	}

	tx, err := dbX.Begin()
	if err != nil {
		return err
	}

	defer tx.Commit()
	//resultsets := make(map[string][]map[string]any, 0)

	//ctx = context.WithValue(ctx, go_ibm_db.ESCAPE_QUOTE, true)  // use strconv.Quote on result set

	rows, err := tx.QueryContext(ctx, preparedCallStatements.FinalCallStatement)

	if err != nil {
		return err
	}
	defer rows.Close()
	resultsets := RowsToResultsets(rows, dummyCall)

	// assign result sets

	for k, v := range resultsets {
		preparedCallStatements.ResponseFormat[k] = v

	}

	for kXX, v := range preparedCallStatements.InOutParamVariables {
		p, found := preparedCallStatements.InOutParamMapToSPParam[kXX]

		if found {
			keyToUse := p.GetNameToUse()

			sqlAgain := fmt.Sprintf("select  @%s;", keyToUse)

			row := tx.QueryRow(sqlAgain)

			err := row.Scan(&v)
			if err != nil {
				v = nil
			}
			// x, ok := (*v).([]byte)
			// if ok {
			// 	*v = x
			// }

			if v != nil && (parameterIsString(p) || reflect.ValueOf(v).Kind() == reflect.String) {

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
				cv, err := ConvertOUTVarToType(p, v)
				if err == nil {
					preparedCallStatements.ResponseFormat[keyToUse] = cv
				} else {
					preparedCallStatements.ResponseFormat[keyToUse] = v
				}

			}

			if p.Mode == "OUT" && keyToUse == "QHTTP_STATUS_CODE" && parameterIsInt(p) {
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

	fmt.Println(qhttp_status_code)

	//	preparedCallStatements.ResponseFormat["data"] = resultsets

	return nil

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MySqlServer) getSPDetails(ctx context.Context, sp *storedProc.StoredProc) error {

	resultSets := 0

	sp.ResultSets = 0

	sqlToRun := ""
	if sp.UseSpecificName {
		sqlToRun = fmt.Sprintf("select trim(ROUTINE_SCHEMA), trim(SPECIFIC_NAME),trim(ROUTINE_SCHEMA),trim(ROUTINE_NAME), 0, SQL_DATA_ACCESS,CREATED from  INFORMATION_SCHEMA.ROUTINES where upper(SPECIFIC_NAME)='%s'  and upper(SPECIFIC_SCHEMA)='%s' and ROUTINE_TYPE='PROCEDURE' limit 1", strings.ToUpper(sp.Name), strings.ToUpper(sp.Lib))
	} else {
		sqlToRun = fmt.Sprintf("select trim(ROUTINE_SCHEMA), trim(SPECIFIC_NAME),trim(ROUTINE_SCHEMA),trim(ROUTINE_NAME), 0, SQL_DATA_ACCESS,CREATED from  INFORMATION_SCHEMA.ROUTINES where upper(ROUTINE_NAME)='%s'  and upper(ROUTINE_SCHEMA)='%s' and ROUTINE_TYPE='PROCEDURE' limit 1", strings.ToUpper(sp.Name), strings.ToUpper(sp.Lib))

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
func (s *MySqlServer) GetParameters(ctx context.Context, sp *storedProc.StoredProc) error {

	originalParams := sp.Parameters

	sqlToUse := fmt.Sprintf("SELECT ORDINAL_POSITION, upper(trim(PARAMETER_MODE)) , upper(trim(PARAMETER_NAME)),DATA_TYPE, ifnull(NUMERIC_SCALE,0), ifnull(NUMERIC_PRECISION,0), ifnull(CHARACTER_MAXIMUM_LENGTH,0),  null FROM INFORMATION_SCHEMA.PARAMETERS WHERE upper(SPECIFIC_NAME)='%s' and   upper(SPECIFIC_SCHEMA) ='%s' ORDER BY ORDINAL_POSITION", strings.ToUpper(sp.SpecificName), strings.ToUpper(sp.SpecificLib))

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
