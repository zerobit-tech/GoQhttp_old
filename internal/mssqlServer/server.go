package mssqlserver

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

type MSSqlServer struct {
	*dbserver.Server
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MSSqlServer) hasSPUpdated(ctx context.Context, sp *storedProc.StoredProc) bool {

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
func (s *MSSqlServer) buildCallStatement(sp *storedProc.StoredProc, useNamedParams bool) (err error) {

	sp.CallStatement = fmt.Sprintf("%s.%s", sp.SpecificLib, sp.SpecificName)
	return nil

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MSSqlServer) prepareCallStatement(sp *storedProc.StoredProc, givenParams map[string]any) (*storedProc.PreparedCallStatements, error) {

	spResponseFormat := make(map[string]any)
	inoutParams := make([]any, 0) // any should be sql.NamedArg
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

			}

			param := sql.Named(p.Name, valueToUse)

			inoutParams = append(inoutParams, param)

		case "INOUT", "OUT":
			spResponseFormat[p.GetNameToUse()] = p.Datatype

			valueToUse, found := givenParams[paramNameToUse]
			if !found {
				valueToUse = s.GetDefaultValue(p)
				if valueToUse == "NULL" {
					valueToUse = nil
				}

			}
			// if !p.HasValidValue(valueToUse) {
			// 	return nil, fmt.Errorf("%s: invalid value", stringutils.AsString(valueToUse))
			// }

			valueToUse, err := p.ConvertToType(valueToUse)
			if err != nil {
				return nil, fmt.Errorf("%s: invalid value", stringutils.AsString(valueToUse))

			}
			inoutParamVariables[p.Name] = &valueToUse

			param := sql.Named(p.Name, sql.Out{Dest: inoutParamVariables[p.Name], In: true})

			inoutParams = append(inoutParams, param)

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
func (s *MSSqlServer) call(ctx context.Context, sp *storedProc.StoredProc, givenParams map[string]any) (*storedProc.StoredProcResponse, time.Duration, error) {
	//log.Printf("%v: %v\n", "SeversCall005.002", time.Now())
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	qhttp_status_code := 200
	qhttp_status_message := ""

	logEntries := make([]*logger.LogEvent, 0)
	preparedCallStatements, err := s.prepareCallStatement(sp, givenParams)
	if err != nil {
		//logEntries = append(logEntries, storedProc.LogByType{Text: err.Error(), Type: "E"})
		return &storedProc.StoredProcResponse{LogData: logEntries}, 0, err
	}

	t1 := time.Now()
	//logEntries = append(logEntries, storedProc.LogByType{Text: "Starting DB CALL", Type: "I"})
	err = s.seversCall(ctx, sp, preparedCallStatements, false)

	spCallDuration := time.Since(t1)

	//logEntries = append(logEntries, storedProc.LogByType{Text: fmt.Sprintf("Finished DB CALL in: %s", spCallDuration), Type: "I"})

	//log.Printf("%v: %v\n", "SeversCall005.004", time.Now())

	if err != nil {
		//logEntries = append(logEntries, storedProc.LogByType{Text: err.Error(), Type: "E"})

		return &storedProc.StoredProcResponse{LogData: logEntries}, 0, err
	}
	//logEntries = append(logEntries, storedProc.LogByType{Text: "SP Call complete", Type: "I"})

	// read INOUT and OUT parameter values
	for kXX, v := range preparedCallStatements.InOutParamVariables {
		p, found := preparedCallStatements.InOutParamMapToSPParam[kXX]

		keyToUse := p.GetNameToUse()

		if found {

			if parameterIsString(p) || reflect.ValueOf(v).Kind() == reflect.String {

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
func (s *MSSqlServer) seversCall(ctx context.Context, sp *storedProc.StoredProc, preparedCallStatements *storedProc.PreparedCallStatements, dummyCall bool) (ferr error) {

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

	//ctx = context.WithValue(ctx, go_ibm_db.ESCAPE_QUOTE, true)  // use strconv.Quote on result set

	rows, err := db.QueryContext(ctx, sp.CallStatement, preparedCallStatements.InOutParams...)

	if err != nil {
		return err
	}
	defer rows.Close()
	resultsets := rowsToResultsets(rows, dummyCall)

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
func (s *MSSqlServer) getSPDetails(ctx context.Context, sp *storedProc.StoredProc) error {

	spCatalog, spSchema := breakCatalogSchema(sp)

	resultSets := 0

	sp.ResultSets = 0

	sqlToRun := ""
	if sp.UseSpecificName {
		sqlToRun = fmt.Sprintf("select top 1 CONCAT(SPECIFIC_CATALOG,'.',SPECIFIC_SCHEMA), trim(SPECIFIC_NAME), CONCAT(ROUTINE_CATALOG,'.',ROUTINE_SCHEMA),trim(ROUTINE_NAME), 0, SQL_DATA_ACCESS,CREATED from %s.INFORMATION_SCHEMA.ROUTINES where SPECIFIC_CATALOG='%s'  and  SPECIFIC_NAME='%s'  and SPECIFIC_SCHEMA='%s' and ROUTINE_TYPE = N'PROCEDURE' ", spCatalog, spCatalog, strings.ToUpper(sp.Name), spSchema)
	} else {
		sqlToRun = fmt.Sprintf("select top 1 CONCAT(SPECIFIC_CATALOG,'.',SPECIFIC_SCHEMA), trim(SPECIFIC_NAME), CONCAT(ROUTINE_CATALOG,'.',ROUTINE_SCHEMA),trim(ROUTINE_NAME), 0, SQL_DATA_ACCESS,CREATED from %s.INFORMATION_SCHEMA.ROUTINES where ROUTINE_CATALOG='%s'  and  ROUTINE_NAME='%s'  and ROUTINE_SCHEMA='%s'  and ROUTINE_TYPE = N'PROCEDURE'  ", spCatalog, spCatalog, strings.ToUpper(sp.Name), spSchema)

	}

	conn, err := s.GetConnection()

	if err != nil {
		return err
	}
	row := conn.QueryRowContext(ctx, sqlToRun)

	err = row.Scan(&sp.SpecificLib, &sp.SpecificName, &sp.Lib, &sp.Name, &resultSets, &sp.DataAccess, &sp.Modified)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("not found")
		}
		return err

	}
	sp.ResultSets = resultSets

	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *MSSqlServer) getParameters(ctx context.Context, sp *storedProc.StoredProc) error {

	spCatalog, _ := breakCatalogSchema(sp)

	originalParams := sp.Parameters

	columnes := "SELECT parameter_id,  cast(is_output as char(1)), upper(trim(p.name)),TYPE_NAME(p.user_type_id) AS parameter_type  , isnull(scale,0), isnull([precision],0), isnull(max_length,0),  default_value "

	tables := fmt.Sprintf("FROM %s.sys.objects AS o INNER JOIN %s.sys.parameters AS p ON o.object_id = p.object_id", spCatalog, spCatalog)

	where := fmt.Sprintf(" WHERE o.object_id = OBJECT_ID('%s.%s') ", sp.SpecificLib, sp.SpecificName)

	orderby := "ORDER BY  p.parameter_id"

	sqlToUse := fmt.Sprintf("%s %s %s %s;", columnes, tables, where, orderby)

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

		if spParamter.Mode == "0" {
			spParamter.Mode = "IN"
		} else {
			spParamter.Mode = "INOUT"
		}

		if strings.TrimSpace(spParamter.Datatype) == "" {
			spParamter.Datatype = "CHARACTER"
		}

		if strings.TrimSpace(spParamter.Name) == "" {
			sp.UseNamedParams = false
			spParamter.Name = strconv.Itoa(spParamter.Position)

		}
		if strings.HasPrefix(strings.TrimSpace(spParamter.Name), "@") {
			spParamter.Name = strings.TrimPrefix(strings.TrimSpace(spParamter.Name), "@")
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
