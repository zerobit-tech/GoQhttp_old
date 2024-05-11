package ibmiServer

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/onlysumitg/godbc"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	"github.com/zerobit-tech/GoQhttp/internal/validator"
	"github.com/zerobit-tech/GoQhttp/logger"
	"github.com/zerobit-tech/GoQhttp/utils/httputils"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
)

type Server struct {
	Mux sync.Mutex `json:"-" db:"-" form:"-"`

	ID string `json:"id" db:"id" form:"id"`

	Name string `json:"server_name" db:"server_name" form:"name"`
	IP   string `json:"ip" db:"ip" form:"ip"`
	Port uint16 `json:"port" db:"port" form:"port"`
	Ssl  bool   `json:"ssl" db:"ssl" form:"ssl"`

	UserName string `json:"un" db:"un" form:"user_name"`
	Password string `json:"pwd" db:"pwd" form:"password"`

	//WorkLib           string    `json:"wlib" db:"wlib" form:"worklib"`
	CreatedAt       time.Time `json:"c_at" db:"c_at" form:"-"`
	UpdatedAt       time.Time `json:"u_at" db:"u_at" form:"-"`
	ConnectionsOpen int       `json:"conn" db:"conn" form:"connections"`
	ConnectionsIdle int       `json:"iconn" db:"iconn" form:"idleconnections"`

	ConnectionMaxAge  int    `json:"cage" db:"cage" form:"cage"`
	ConnectionIdleAge int    `json:"icage" db:"icage" form:"icage"`
	PingTimeout       int    `json:"pingtout" db:"pingtout" form:"pingtout"`
	PingQuery         string `json:"pingquery" db:"pingquery" form:"pingquery"`

	OnHold        bool   `json:"oh" db:"oh" form:"onhold"`
	OnHoldMessage string `json:"ohm" db:"ohm" form:"onholdmessage"`

	ConfigFileLib string `json:"configfilelib" db:"configfilelib" form:"configfilelib"`
	ConfigFile    string `json:"configfile" db:"configfile" form:"configfile"`

	AutoPromotePrefix string `json:"autopromoteprefix" db:"autopromoteprefix" form:"autopromoteprefix"`

	UserTokenFileLib string `json:"usertokenfilelib" db:"usertokenfilelib" form:"usertokenfilelib"`
	UserTokenFile    string `json:"usertokenfile" db:"usertokenfile" form:"usertokenfile"`

	LibList []string `json:"liblist" db:"liblist" form:"liblist"`

	LastAutoPromoteDate string `json:"lastautopromotecheck" db:"lastautopromotecheck" form:"lastautopromotecheck"`

	//Namespace string `json:"namespace" db:"namespace" form:"namespace"`
	validator.Validator `json:"-" db:"-" form:"-"`
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) PromoationTableQuery() []string {

	tableNametoUse := s.ConfigFile
	tableLibNametoUse := s.ConfigFileLib

	if tableNametoUse == "" {
		tableNametoUse = "{{ENTER_TABLE_NAME}}"
	}

	if tableLibNametoUse == "" {
		tableLibNametoUse = "{{ENTER_TABLE_LIB_NAME}}"
	}

	return []string{
		fmt.Sprintf("create or replace table %s.%s", tableLibNametoUse, tableNametoUse),
		"(",
		"operation char(1) NOT NULL WITH DEFAULT,",
		"endpoint varchar(100) NOT NULL WITH DEFAULT,",
		"namespace varchar(100) NOT NULL WITH DEFAULT,",
		"storedproc varchar(256) NOT NULL WITH DEFAULT,",
		"storedproclib varchar(100) NOT NULL WITH DEFAULT,",
		"httpmethod char(10) NOT NULL WITH DEFAULT,",
		"usespecificname char(1) NOT NULL WITH DEFAULT,",
		"usewithoutauth char(1) NOT NULL WITH DEFAULT,",
		"paramalias varchar(500) NOT NULL WITH DEFAULT,",
		"paramplacement varchar(500) NOT NULL WITH DEFAULT,",
		"status char(1) NOT NULL WITH DEFAULT,",
		"statusmessage varchar(100) NOT NULL WITH DEFAULT",

		")",
	}
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) CreatePromotionTable(ctx context.Context) error {

	if s.ConfigFile == "" || s.ConfigFileLib == "" {

		return fmt.Errorf("Error: Promotion Table or LIB not given for %s", s.Name)

	}

	sqlToRun := strings.Join(s.PromoationTableQuery(), " ")
	conn, err := s.GetSingleConnection()

	if err != nil {
		log.Println("Error CreatePromotionTable:", err.Error())
		return err
	}
	_, err = conn.ExecContext(ctx, sqlToRun)

	if err != nil {
		log.Println("Error CreatePromotionTable:", err.Error(), sqlToRun)

		return err

	}

	return nil
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Server) LogImage() string {
	imageMap := make(map[string]any)
	imageMap["Name"] = s.Name
	imageMap["IP"] = s.IP
	imageMap["UserName"] = s.UserName

	imageMap["ConnectionsOpen"] = s.ConnectionsOpen
	imageMap["ConnectionsIdle"] = s.ConnectionsIdle
	imageMap["ConnectionMaxAge"] = s.ConnectionMaxAge
	imageMap["ConnectionIdleAge"] = s.ConnectionIdleAge

	imageMap["ConfigFileLib"] = s.ConfigFileLib
	imageMap["ConfigFile"] = s.ConfigFile

	imageMap["AutoPromotePrefix"] = s.AutoPromotePrefix

	imageMap["UserTokenFileLib"] = s.UserTokenFileLib
	imageMap["UserTokenFile"] = s.UserTokenFile

	imageMap["LibList"] = s.LibList

	j, err := json.MarshalIndent(imageMap, " ", " ")
	if err == nil {
		return string(j)
	}

	return err.Error()
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *Server) hasSPUpdated(ctx context.Context, sp *storedProc.StoredProc) bool {

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
//	https://www.ibm.com/docs/en/i/7.4?topic=details-connection-string-keywords
//	Connection String: DBQ
//
// -----------------------------------------------------------------
func (s *Server) GetConnetionLibList() string {
	libList := "*USRLIBL"
	severLibL := s.GetLibListString()
	if severLibL != "" {
		libList = "," + severLibL + ",*USRLIBL" // DBQ=,mylib,mylib2,mylib3;NAM=1 is specified, then CURRENT SCHEMA special register is set to *LIBL and the library list would be MYLIB, MYLIB2, and MYLIB3.
	}
	return libList
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *Server) GetLibListString() string {
	if len(s.LibList) <= 0 {
		return ""
	}

	libList := make([]string, 0)

	for _, lib := range s.LibList {
		if strings.TrimSpace(lib) != "" {
			libList = append(libList, strings.ToUpper(lib))
		}
	}
	libListString := strings.Join(libList, ",")

	return libListString

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *Server) ManageLibList() {
	libList := make([]string, 20)
	iCounter := 0
	for _, lib := range s.LibList {
		if strings.TrimSpace(lib) != "" {
			libList[iCounter] = strings.ToUpper(lib)
			iCounter += 1
		}
	}

	s.LibList = libList

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (s *Server) buildCallStatement(sp *storedProc.StoredProc, useNamedParams bool) (err error) {

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
func (s *Server) prepareCallStatement(sp *storedProc.StoredProc, givenParams map[string]any, regexMap map[string]string) (*storedProc.PreparedCallStatements, error) {

	spResponseFormat := make(map[string]any)
	inoutParams := make([]any, 0)
	inoutParamVariables := make(map[string]*any)
	inOutParamMapToSPParam := make(map[string]*storedProc.StoredProcParamter)

	finalCallStatement := sp.CallStatement

	for _, p := range sp.Parameters {
		paramNameToUse := p.GetNameToUse(false)

		switch p.Mode {
		case "IN":
			valueToUse, found := givenParams[paramNameToUse]
			if !found {
				valueToUse = s.getDefaultValue(p)

			} else {
				//p.GivenValue = asString(valueToUse)

			}

			// parameter validation !!

			if err := p.HasValidValue(valueToUse, regexMap); err != nil {
				return nil, err
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
			spResponseFormat[p.GetNameToUse(true)] = p.Datatype

			valueToUse, found := givenParams[paramNameToUse]
			if !found {
				valueToUse = s.getDefaultValue(p)
				if valueToUse == "NULL" {
					valueToUse = nil
				}

			} else {
				//p.GivenValue = asString(valueToUse)

			}
			if err := p.HasValidValue(valueToUse, regexMap); err != nil {
				return nil, err
			}

			valueToUse, err := p.ConvertToType(valueToUse)
			if err != nil {
				return nil, fmt.Errorf("%s: invalid value", paramNameToUse)

			}
			inoutParamVariables[p.Name] = &valueToUse

			inoutParams = append(inoutParams, sql.Out{Dest: inoutParamVariables[p.Name], In: true})

			inOutParamMapToSPParam[p.Name] = p

		case "OUT":
			spResponseFormat[p.GetNameToUse(true)] = p.Datatype
			out := getParameterofType(p)
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
func (s *Server) call(ctx context.Context, callID string, sp *storedProc.StoredProc, givenParams map[string]any, paramRegex map[string]string) (*storedProc.StoredProcResponse, time.Duration, error) {
	//log.Printf("%v: %v\n", "SeversCall005.002", time.Now())
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	qhttp_status_code := 200
	qhttp_status_message := ""

	logEntries := make([]*logger.LogEvent, 0, 5)
	preparedCallStatements, err := s.prepareCallStatement(sp, givenParams, paramRegex)
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

		keyToUse := p.GetNameToUse(true)

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
						} else {
							jsonDataList := make([]any, 0)
							err := json.Unmarshal(b, &jsonDataList)
							if err == nil {
								preparedCallStatements.ResponseFormat[keyToUse] = &jsonDataList
								assignStrVal = false
							}
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

				httpCode, message := httputils.GetValidHttpCode(*v)

				if httpCode > 0 {
					qhttp_status_code = httpCode

					// remove QHTTP_STATUS_CODE from out params
					delete(preparedCallStatements.ResponseFormat, keyToUse)

					if qhttp_status_message == "" {
						qhttp_status_message = message
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
func (s *Server) SeversCall(ctx context.Context, sp *storedProc.StoredProc, preparedCallStatements *storedProc.PreparedCallStatements, dummyCall bool) (ferr error) {

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
	ctx = context.WithValue(ctx, godbc.LOAD_SP_RESULT_SETS, resultsets)
	ctx = context.WithValue(ctx, godbc.DUMMY_SP_CALL, dummyCall)
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
func (s *Server) getResultSetCount(ctx context.Context, sp *storedProc.StoredProc) error {

	resultSets := 0

	sp.ResultSets = 0

	sqlToRun := ""
	if sp.UseSpecificName {
		sqlToRun = fmt.Sprintf("select trim(SPECIFIC_SCHEMA), trim(SPECIFIC_NAME),trim(ROUTINE_SCHEMA),trim(ROUTINE_NAME), ifnull(RESULT_SETS,0), SQL_DATA_ACCESS,ROUTINE_CREATED from qsys2.sysprocs where SPECIFIC_NAME='%s'  and SPECIFIC_SCHEMA='%s' limit 1", strings.ToUpper(sp.Name), strings.ToUpper(sp.Lib))
	} else {
		sqlToRun = fmt.Sprintf("select trim(SPECIFIC_SCHEMA), trim(SPECIFIC_NAME),trim(ROUTINE_SCHEMA),trim(ROUTINE_NAME), ifnull(RESULT_SETS,0), SQL_DATA_ACCESS,ROUTINE_CREATED from qsys2.sysprocs where ROUTINE_NAME='%s'  and ROUTINE_SCHEMA='%s' limit 1", strings.ToUpper(sp.Name), strings.ToUpper(sp.Lib))

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
func (s *Server) getParameters(ctx context.Context, sp *storedProc.StoredProc) error {

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
				p.Placement = op.Placement
			}
		}
	}

	sp.AssignAliasForPathPlacement()

	return nil

}
