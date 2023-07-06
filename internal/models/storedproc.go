package models

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

	"github.com/gosimple/slug"
	"github.com/onlysumitg/GoQhttp/go_ibm_db"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/utils/httputils"
	bolt "go.etcd.io/bbolt"
)

type LogByType struct {
	Text string `json:"-" db:"-" form:"-"`
	Type string `json:"-" db:"-" form:"-"`
}

type StoredProcResponse struct {
	ReferenceId string
	Status      int
	Message     string
	Data        map[string]any
	LogData     []LogByType `json:"-" db:"-" form:"-"`
}

type ServerRecord struct {
	ID   string `json:"id" db:"id" form:"id"`
	Name string `json:"server_name" db:"server_name" form:"name"`
}

type StoredProc struct {
	ID           string `json:"id" db:"id" form:"id"`
	EndPointName string `json:"endpointname" db:"endpointname" form:"endpointname"`
	HttpMethod   string `json:"httpmethod" db:"httpmethod" form:"httpmethod"`

	Name            string                `json:"name" db:"name" form:"name"`
	Lib             string                `json:"lib" db:"lib" form:"lib"`
	SpecificName    string                `json:"specificname" db:"specificname" form:"specificname"`
	SpecificLib     string                `json:"specificlib" db:"specificlib" form:"specificlib"`
	UseSpecificName bool                  `json:"usespecificname" db:"usespecificname" form:"usespecificname"`
	CallStatement   string                `json:"callstatement" db:"callstatement" form:"-"`
	Parameters      []*StoredProcParamter `json:"params" db:"params" form:"-"`
	ResultSets      int                   `json:"resultsets" db:"resultsets" form:"-"`
	ResponseFormat  string                `json:"responseformat" db:"responseformat" form:"-"`

	DefaultServerId    string          `json:"-" db:"-" form:"serverid"`
	DefaultServer      *ServerRecord   `json:"dserverid" db:"dserverid" form:"-"`
	AllowedOnServers   []*ServerRecord `json:"allowedonservers" db:"allowedonservers" form:"allowedonservers"`
	MockUrl            string          `json:"mockurl" db:"mockurl" form:"-"`
	MockUrlWithoutAuth string          `json:"mockurlnoa" db:"mockurlnoa" form:"-"`

	InputPayload        string                     `json:"inputpayload" db:"inputpayload" form:"inputpayload"`
	validator.Validator `json:"-" db:"-" form:"-"` // this contains the fielderror

	AllowWithoutAuth bool `json:"awoauth" db:"awoauth" form:"awoauth"`

	DataAccess string `json:"dataaccess" db:"dataaccess" form:"dataaccess"`

	Modified string `json:"modified" db:"modified" form:"modified"`

	UseNamedParams bool `json:"useunnamedparams" db:"useunnamedparams" form:"-"`

	Promotionsql string `json:"promotionsql" db:"promotionsql" form:"-"`
}

type PreparedCallStatements struct {
	ResponseFormat         map[string]any
	InOutParams            []any
	InOutParamVariables    map[string]*any
	InOutParamMapToSPParam map[string]*StoredProcParamter
	FinalCallStatement     string
}

// ------------------------------------------------------------
// BuildMockUrl(s)
// ------------------------------------------------------------
func (s *StoredProc) Slug() string {
	return slug.Make(s.EndPointName + "_" + s.HttpMethod)

}

// ------------------------------------------------------------
// BuildMockUrl(s)
// ------------------------------------------------------------
func (s *StoredProc) ValidateAlias() error {

	for _, p1 := range s.Parameters {
		for _, p2 := range s.Parameters {
			if p1.Name != p2.Name && p1.GetNameToUse() == p2.GetNameToUse() {
				return fmt.Errorf("Conflict between %s and %s.", p1.Name, p2.Name)
			}
		}

	}
	return nil
}

// ------------------------------------------------------------
// BuildMockUrl(s)
// ------------------------------------------------------------
func (s *StoredProc) IsAllowedForServer(server *Server) bool {
	if server == nil {
		return false
	}

	for _, rcd := range s.AllowedOnServers {
		if server.ID == rcd.ID {
			return true
		}
	}

	return false

}

// ------------------------------------------------------------
// BuildMockUrl(s)
// ------------------------------------------------------------
func (s *StoredProc) AddAllowedServer(server *Server) {
	alreadyAssigned := false

	for _, rcd := range s.AllowedOnServers {
		if server.ID == rcd.ID {
			alreadyAssigned = true
			rcd.Name = server.Name
		}
	}

	if !alreadyAssigned {
		rcd := &ServerRecord{ID: server.ID, Name: server.Name}
		s.AllowedOnServers = append(s.AllowedOnServers, rcd)
	}

}

// ------------------------------------------------------------
// BuildMockUrl(s)
// ------------------------------------------------------------
func (s *StoredProc) DeleteAllowedServer(server *Server) {

	a := make([]*ServerRecord, 0)

	for _, rcd := range s.AllowedOnServers {
		if server.ID != rcd.ID {
			a = append(a, rcd)
		}
	}

	s.AllowedOnServers = a

}

// ------------------------------------------------------------
// BuildMockUrl(s)
// ------------------------------------------------------------
func (s *StoredProc) BuildMockUrl() {

	queryParamString := ""
	inputPayload := make(map[string]string)

outerloop:
	for _, p := range s.Parameters {
		if p.Mode == "OUT" {
			continue outerloop
		}

		// dont display inbuilt param
		for _, ibp := range InbuiltParams {
			if strings.EqualFold(ibp, p.GetNameToUse()) {
				continue outerloop
			}
		}

		inputPayload[p.GetNameToUse()] = fmt.Sprintf("{%s}", p.Datatype)
		if queryParamString == "" {
			queryParamString = fmt.Sprintf("?%s={%s}", p.GetNameToUse(), p.Datatype)
		} else {
			queryParamString = queryParamString + fmt.Sprintf("&%s={%s}", p.GetNameToUse(), p.Datatype)

		}

	}

	if s.HttpMethod != "GET" {

		jsonPayload, err := json.MarshalIndent(inputPayload, "", "  ")
		if err == nil {
			s.InputPayload = string(jsonPayload)
		}
		queryParamString = ""
	}

	s.MockUrl = fmt.Sprintf("api/%s%s", s.EndPointName, queryParamString)
	s.MockUrlWithoutAuth = fmt.Sprintf("uapi/%s%s", s.EndPointName, queryParamString)
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (sp *StoredProc) Refresh(ctx context.Context, s *Server) error {
	if sp.HasSPUpdated(ctx, s) {
		err := sp.PreapreToSave(ctx, *s)
		if err != nil {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (sp *StoredProc) HasSPUpdated(ctx context.Context, s *Server) bool {

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
func (sp *StoredProc) Exists(ctx context.Context, s *Server) (bool, error) {

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
func (sp *StoredProc) PreapreToSave(ctx context.Context, s Server) error {
	sp.Name = strings.ToUpper(strings.TrimSpace(sp.Name))
	sp.Lib = strings.ToUpper(strings.TrimSpace(sp.Lib))
	sp.HttpMethod = strings.ToUpper(strings.TrimSpace(sp.HttpMethod))
	sp.UseNamedParams = true

	ctx1, cancelFunc1 := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc1()
	err := sp.GetResultSetCount(ctx1, &s)
	if err != nil {
		return err
	}

	ctx2, cancelFunc2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc2()
	err = sp.GetParameters(ctx2, &s)
	if err != nil {
		return err
	}

	for _, p := range sp.Parameters {
		if IsUnsupportedDataType(p.Datatype, p.Mode) {
			return fmt.Errorf("%s %s (datatype %s) not supported", p.Mode, p.Name, p.Datatype)
		}
	}

	sp.buildCallStatement(sp.UseNamedParams)
	sp.BuildMockUrl()

	sp.BuildPromotionSQL(&s)

	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (sp *StoredProc) buildCallStatement(useNamedParams bool) (err error) {

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
func (sp *StoredProc) prepareCallStatement(s *Server, givenParams map[string]any) (*PreparedCallStatements, error) {

	spResponseFormat := make(map[string]any)
	inoutParams := make([]any, 0)
	inoutParamVariables := make(map[string]*any)
	inOutParamMapToSPParam := make(map[string]*StoredProcParamter)

	finalCallStatement := sp.CallStatement

	for _, p := range sp.Parameters {
		paramNameToUse := p.GetNameToUse()

		switch p.Mode {
		case "IN":
			valueToUse, found := givenParams[paramNameToUse]
			if !found {
				valueToUse = p.GetDefaultValue(s)

			} else {
				//p.GivenValue = asString(valueToUse)

			}
			if !p.HasValidValue(valueToUse) {
				return nil, fmt.Errorf("%s: invalid value", p.Name)
			}

			stringToReplace := ""
			if p.NeedQuote(asString(valueToUse)) {
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
				valueToUse = p.GetDefaultValue(s)
				if valueToUse == "NULL" {
					valueToUse = nil
				}

			} else {
				//p.GivenValue = asString(valueToUse)

			}
			if !p.HasValidValue(valueToUse) {
				return nil, fmt.Errorf("%s: invalid value", asString(valueToUse))
			}

			valueToUse, err := p.ConvertToType(valueToUse)
			if err != nil {
				return nil, fmt.Errorf("%s: invalid value", asString(valueToUse))

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

	return &PreparedCallStatements{
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
func (sp *StoredProc) APICall(ctx context.Context, s *Server, apiCall *ApiCall) {
	//log.Printf("%v: %v\n", "SeversCall005.001", time.Now())

	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	defer func() {
		if r := recover(); r != nil {
			responseFormat := &StoredProcResponse{
				ReferenceId: "string",
				Status:      500,
				Message:     fmt.Sprintf("%s", r),
				Data:        map[string]any{},
				LogData:     []LogByType{{Text: fmt.Sprintf("%s", r), Type: "ERROR"}},
			}
			apiCall.Response = responseFormat
			apiCall.Err = fmt.Errorf("%s", r)
		}
	}()

	givenParams := make(map[string]any)
	apiCall.LogInfo("Building parameters for SP call")
	for k, v := range apiCall.RequestFlatMap {
		givenParams[k] = v.Value
	}
	apiCall.Response, apiCall.SPCallDuration, apiCall.Err = sp.Call(ctx, s, givenParams)

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (sp *StoredProc) Call(ctx context.Context, s *Server, givenParams map[string]any) (*StoredProcResponse, time.Duration, error) {
	//log.Printf("%v: %v\n", "SeversCall005.002", time.Now())
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	qhttp_status_code := 200
	qhttp_status_message := ""

	logEntries := make([]LogByType, 0)
	preparedCallStatements, err := sp.prepareCallStatement(s, givenParams)
	if err != nil {
		logEntries = append(logEntries, LogByType{Text: err.Error(), Type: "E"})
		return &StoredProcResponse{LogData: logEntries}, 0, err
	}

	t1 := time.Now()
	logEntries = append(logEntries, LogByType{Text: "Starting DB CALL", Type: "I"})
	err = sp.SeversCall(ctx, s, preparedCallStatements, false)

	spCallDuration := time.Since(t1)

	logEntries = append(logEntries, LogByType{Text: fmt.Sprintf("Finished DB CALL in: %s", spCallDuration), Type: "I"})

	//log.Printf("%v: %v\n", "SeversCall005.004", time.Now())

	if err != nil {
		logEntries = append(logEntries, LogByType{Text: err.Error(), Type: "E"})

		return &StoredProcResponse{LogData: logEntries}, 0, err
	}
	logEntries = append(logEntries, LogByType{Text: "SP Call complete", Type: "I"})

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

	responseFormat := &StoredProcResponse{
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
func (sp *StoredProc) DummyCall(s *Server, givenParams map[string]any) (*StoredProcResponse, error) {
	preparedCallStatements, err := sp.prepareCallStatement(s, givenParams)
	if err != nil {
		return nil, err
	}
	err = sp.SeversCall(context.Background(), s, preparedCallStatements, true)
	if err != nil {
		return nil, err
	}

	responseFormat := &StoredProcResponse{
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
func (sp *StoredProc) SeversCall(ctx context.Context, s *Server, preparedCallStatements *PreparedCallStatements, dummyCall bool) (ferr error) {

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
func (sp *StoredProc) GetResultSetCount(ctx context.Context, s *Server) error {

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
			return SpNotFound
		}
		return err

	}
	sp.ResultSets = resultSets

	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (sp *StoredProc) GetParameters(ctx context.Context, s *Server) error {

	originalParams := sp.Parameters

	sqlToUse := fmt.Sprintf("SELECT ORDINAL_POSITION, upper(trim(PARAMETER_MODE)) , upper(trim(PARAMETER_NAME)),DATA_TYPE, ifnull(NUMERIC_SCALE,0), ifnull(NUMERIC_PRECISION,0), ifnull(CHARACTER_MAXIMUM_LENGTH,0),  default FROM qsys2.sysparms WHERE SPECIFIC_NAME='%s' and   SPECIFIC_SCHEMA ='%s' ORDER BY ORDINAL_POSITION", strings.ToUpper(sp.SpecificName), strings.ToUpper(sp.SpecificLib))

	sp.Parameters = make([]*StoredProcParamter, 0)
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
		spParamter := &StoredProcParamter{}
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

// // -----------------------------------------------------------------
// //
// // -----------------------------------------------------------------
// func (s *SavedQuery) PopulateFieldsXX() {
// 	s.Fields = findQueryFields(s.Sql)
// }

// // -----------------------------------------------------------------
// //
// // -----------------------------------------------------------------
// func (s *SavedQuery) ReplaceFieldsXX(values map[string]string) (string, map[string]string) {

// 	return ReplaceQueryFields(s.Sql, values)
// }

// // -----------------------------------------------------------------
// //
// // -----------------------------------------------------------------
// func ReplaceQueryFieldsXX(sqlString string, values map[string]string) (string, map[string]string) {

// 	fieldErrors := make(map[string]string)
// 	fields := findQueryFields(sqlString)
// 	sql := sqlString
// 	log.Println("sql1>>>>>", sql, fields)

// 	for _, field := range fields {
// 		fieldValue, found := values[field.Name]
// 		if found {
// 			sql = strings.ReplaceAll(sql, field.ID, fieldValue)
// 			log.Println("sql>>>>>", sql)
// 		} else if field.DefaultValue != "" {
// 			sql = strings.ReplaceAll(sql, field.ID, field.DefaultValue)
// 		} else {
// 			fieldErrors[field.Name] = "Field value is required"
// 		}

// 	}
// 	return sql, fieldErrors
// }

// // -----------------------------------------------------------------
// //
// // -----------------------------------------------------------------
// func findQueryFieldsXX(str string) []*QueryField {

// 	var re = regexp.MustCompile(`(?m)({{.*?}})`)

// 	fields := make([]*QueryField, 0)
// 	fieldNames := make([]string, 0)

// 	for _, match := range re.FindAllString(str, -1) {
// 		field := fieldToQueryField(match)

// 		if !isInList(fieldNames, field.Name) { // not found
// 			fieldNames = append(fieldNames, field.Name)
// 			fields = append(fields, field)
// 		}

// 		//fmt.Println(match, "found at index", i)
// 	}
// 	return fields
// }

// // -----------------------------------------------------------------
// //
// //	TODO --> improve search
// //
// // -----------------------------------------------------------------
// func isInListXX(list []string, search string) bool {
// 	for _, val := range list {
// 		if strings.EqualFold(val, search) {
// 			return true
// 		}
// 	}

// 	return false
// }

// // -----------------------------------------------------------------
// //
// // -----------------------------------------------------------------

// func fieldToQueryFieldXX(str string) *QueryField {
// 	field := strings.TrimRight(str, "}")
// 	field = strings.TrimLeft(field, "{")

// 	fieldNameValue := strings.Split(field, ":")

// 	queryField := QueryField{ID: str}
// 	queryField.Name = strings.Trim(fieldNameValue[0], " ")

// 	if len(fieldNameValue) > 1 {
// 		queryField.DefaultValue = fieldNameValue[1]
// 	}

// 	return &queryField
// }

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Define a new UserModel type which wraps a database connection pool.
type StoredProcModel struct {
	DB *bolt.DB
}

func (m *StoredProcModel) getTableName() []byte {
	return []byte("storedprocs")
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *StoredProcModel) Save(u *StoredProc) (string, error) {
	var id string
	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getTableName())
		if err != nil {
			return err
		}
		u.Name = strings.ToUpper(strings.TrimSpace(u.Name))

		// generate new ID if id is blank else use the old one to update
		if u.ID == "" {
			u.ID = u.Slug() //uuid.NewString()
			//u.AllowedOnServers = make([]*ServerRecord, 0)
		}
		u.Name = strings.ToUpper(strings.TrimSpace(u.Name))
		u.Lib = strings.ToUpper(strings.TrimSpace(u.Lib))
		u.EndPointName = strings.ToLower(strings.TrimSpace(u.EndPointName))
		id = u.ID
		// Marshal user data into bytes.
		buf, err := json.Marshal(u)
		if err != nil {
			return err
		}

		// key = > user.name+ user.id
		key := strings.ToUpper(u.ID) // + string(itob(u.ID))

		return bucket.Put([]byte(key), buf)
	})

	return id, err
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *StoredProcModel) Delete(id string) error {

	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getTableName())
		if err != nil {
			return err
		}
		key := strings.ToUpper(id)
		dbDeleteError := bucket.Delete([]byte(key))
		return dbDeleteError
	})

	return err
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Insert method to add a new record to the "users" table.
func (m *StoredProcModel) DeleteByName(name string, method string) error {

	for _, sp := range m.List() {
		if strings.EqualFold(sp.EndPointName, name) && strings.EqualFold(sp.HttpMethod, method) {
			err := m.Delete(sp.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *StoredProcModel) Exists(id string) bool {

	var userJson []byte

	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.getTableName())
		if err != nil {
			return err
		}
		key := strings.ToUpper(id)

		userJson = bucket.Get([]byte(key))

		return nil

	})

	return (userJson != nil)
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *StoredProcModel) Duplicate(u *StoredProc) bool {
	exists := false
	for _, sp := range m.List() {

		if sp.ID != u.ID && strings.EqualFold(sp.EndPointName, u.EndPointName) && strings.EqualFold(sp.HttpMethod, u.HttpMethod) {
			exists = true
			break
		}
	}

	return exists
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *StoredProcModel) Get(id string) (*StoredProc, error) {

	if id == "" {
		return nil, errors.New("SavedQuery blank id not allowed")
	}
	var savedQueryJSON []byte // = make([]byte, 0)

	err := m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		savedQueryJSON = bucket.Get([]byte(strings.ToUpper(id)))

		return nil

	})
	savedQuery := StoredProc{}
	if err != nil {
		return &savedQuery, err
	}

	// log.Println("savedQueryJSON >2 >>", savedQueryJSON)

	if savedQueryJSON != nil {
		err := json.Unmarshal(savedQueryJSON, &savedQuery)
		return &savedQuery, err

	}

	return &savedQuery, ErrSavedQueryNotFound

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (m *StoredProcModel) List() []*StoredProc {
	savedQueries := make([]*StoredProc, 0)
	_ = m.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.getTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			savedQuery := StoredProc{}
			err := json.Unmarshal(v, &savedQuery)
			if err == nil {
				savedQueries = append(savedQueries, &savedQuery)
			}
		}

		return nil
	})
	return savedQueries

}
