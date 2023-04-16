package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/utils/timeutils"
	bolt "go.etcd.io/bbolt"
)

type StoredProcResponse struct {
	ReferenceId string
	Status      int
	Message     string
	Data        map[string]any
}

type StoredProc struct {
	ID           string `json:"id" db:"id" form:"id"`
	EndPointName string `json:"endpointname" db:"endpointname" form:"endpointname"`
	HttpMethod   string `json:"httpmethod" db:"httpmethod" form:"httpmethod"`

	Name                string                     `json:"name" db:"name" form:"name"`
	Lib                 string                     `json:"lib" db:"lib" form:"lib"`
	SpecificName        string                     `json:"specificname" db:"specificname" form:"specificname"`
	SpecificLib         string                     `json:"specificlib" db:"specificlib" form:"specificlib"`
	UseSpecificName     bool                       `json:"usespecificname" db:"usespecificname" form:"usespecificname"`
	CallStatement       string                     `json:"callstatement" db:"callstatement" form:"-"`
	Parameters          []*StoredProcParamter      `json:"params" db:"params" form:"-"`
	ResultSets          int                        `json:"resultsets" db:"resultsets" form:"-"`
	validator.Validator `json:"-" db:"-" form:"-"` // this contains the fielderror
	ResponseFormat      string                     `json:"responseformat" db:"responseformat" form:"-"`
}

type PreparedCallStatements struct {
	ResponseFormat         map[string]any
	InOutParams            []any
	InOutParamVariables    map[string]*any
	InOutParamMapToSPParam map[string]*StoredProcParamter
	FinalCallStatement     string
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (sp *StoredProc) PreapreToSave(s Server) error {
	sp.Name = strings.ToUpper(strings.TrimSpace(sp.Name))
	sp.Lib = strings.ToUpper(strings.TrimSpace(sp.Lib))
	err := sp.GetResultSetCount(s)
	if err != nil {
		return err
	}

	err = sp.GetParameters(s)
	if err != nil {
		return err
	}

	sp.buildCallStatement(true)

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
			paramString += fmt.Sprintf("%s %s", value, ",")
		}

	}

	paramString = strings.TrimRight(paramString, ",")
	sp.CallStatement = fmt.Sprintf("call %s.%s(%s)", sp.Lib, sp.Name, paramString)
	return nil

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (sp *StoredProc) prepareCallStatement(givenParams map[string]any) (*PreparedCallStatements, error) {
	spResponseFormat := make(map[string]any)
	inoutParams := make([]any, 0)
	inoutParamVariables := make(map[string]*any)
	inOutParamMapToSPParam := make(map[string]*StoredProcParamter)

	finalCallStatement := sp.CallStatement

	for _, p := range sp.Parameters {
		switch p.Mode {
		case "IN":
			valueToUse, found := givenParams[p.Name]
			if !found {
				valueToUse = p.GetDefaultValue()
			} else {
				p.GivenValue = asString(valueToUse)

			}
			if !p.HasValidValue(valueToUse) {
				return nil, fmt.Errorf("%s: invalid value", p.Name)
			}
			finalCallStatement = strings.ReplaceAll(finalCallStatement, fmt.Sprintf("{:%s}", p.Name), asString(valueToUse))

		case "INOUT":
			spResponseFormat[p.Name] = p.Datatype

			valueToUse, found := givenParams[p.Name]
			if !found {
				valueToUse = p.GetDefaultValue()
			} else {
				p.GivenValue = asString(valueToUse)

			}
			if !p.HasValidValue(valueToUse) {
				return nil, fmt.Errorf("%s: invalid value", asString(valueToUse))
			}

			inoutParamVariables[p.Name] = &valueToUse

			inoutParams = append(inoutParams, sql.Out{Dest: inoutParamVariables[p.Name], In: true})

			inOutParamMapToSPParam[p.Name] = p

		case "OUT":
			spResponseFormat[p.Name] = p.Datatype
			var out any
			inoutParamVariables[p.Name] = &out
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
func (sp *StoredProc) APICall(s Server, apiCall *ApiCall) {
	givenParams := make(map[string]any)

	for k, v := range apiCall.RequestFlatMap {
		givenParams[k] = v.Value
	}
	apiCall.Response, apiCall.Err = sp.Call(s, givenParams)

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (sp *StoredProc) Call(s Server, givenParams map[string]any) (*StoredProcResponse, error) {
	preparedCallStatements, err := sp.prepareCallStatement(givenParams)
	if err != nil {
		return &StoredProcResponse{}, err
	}

	err = sp.SeversCall(s, preparedCallStatements, false)
	if err != nil {
		return &StoredProcResponse{}, err
	}

	// read INOUT and OUT parameter values
	for k, v := range preparedCallStatements.InOutParamVariables {

		p, found := preparedCallStatements.InOutParamMapToSPParam[k]
		if found {
			if p.IsString() {

				preparedCallStatements.ResponseFormat[k] = string((*v).([]byte))

			} else {
				preparedCallStatements.ResponseFormat[k] = v

			}

		}
	}

	responseFormat := &StoredProcResponse{
		ReferenceId: "string",
		Status:      200,
		Message:     "string",
		Data:        preparedCallStatements.ResponseFormat,
	}

	return responseFormat, nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (sp *StoredProc) DummyCall(s Server, givenParams map[string]any) (*StoredProcResponse, error) {
	preparedCallStatements, err := sp.prepareCallStatement(givenParams)
	if err != nil {
		return nil, err
	}
	err = sp.SeversCall(s, preparedCallStatements, true)
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
func (sp *StoredProc) SeversCall(s Server, preparedCallStatements *PreparedCallStatements, dummyCall bool) error {
	defer timeutils.Duration(timeutils.Track("SeversCall"))

	log.Printf("%v: %v\n", "SeversCall 1", time.Now())
	db, err := s.GetConnection()
	if err != nil {
		return err
	}
	log.Printf("%v: %v\n", "SeversCall 2", time.Now())
	resultsets := make(map[string][]map[string]any, 0)
	ctx := context.WithValue(context.Background(), "go_ibm_db_ROW", resultsets)
	ctx = context.WithValue(ctx, "go_ibm_db_DUMMY_CALL", dummyCall)
	_, err = db.ExecContext(ctx, preparedCallStatements.FinalCallStatement, preparedCallStatements.InOutParams...)
	log.Printf("%v: %v\n", "SeversCall 3", time.Now())
	if err != nil {
		return err
	}

	// assign result sets
	preparedCallStatements.ResponseFormat["data"] = resultsets

	return nil

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (sp *StoredProc) GetResultSetCount(s Server) error {

	resultSets := 0

	sp.ResultSets = 0

	sqlToRun := ""
	if sp.UseSpecificName {
		sqlToRun = fmt.Sprintf("select trim(SPECIFIC_SCHEMA), trim(SPECIFIC_NAME),trim(ROUTINE_SCHEMA),trim(ROUTINE_NAME), RESULT_SETS from qsys2.sysprocs where SPECIFIC_NAME='%s'  and SPECIFIC_SCHEMA='%s' limit 1", strings.ToUpper(sp.Name), strings.ToUpper(sp.Lib))
	} else {
		sqlToRun = fmt.Sprintf("select trim(SPECIFIC_SCHEMA), trim(SPECIFIC_NAME),trim(ROUTINE_SCHEMA),trim(ROUTINE_NAME), RESULT_SETS from qsys2.sysprocs where SPECIFIC_NAME='%s'  and SPECIFIC_SCHEMA='%s' limit 1", strings.ToUpper(sp.Name), strings.ToUpper(sp.Lib))

	}

	conn, err := s.GetConnection()

	if err != nil {
		return err
	}
	row := conn.QueryRow(sqlToRun)

	err = row.Scan(&sp.SpecificLib, &sp.SpecificName, &sp.Lib, &sp.Name, &resultSets)

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
func (sp *StoredProc) GetParameters(s Server) error {

	sql := fmt.Sprintf("SELECT ORDINAL_POSITION, upper(trim(PARAMETER_MODE)) , upper(trim(PARAMETER_NAME)),DATA_TYPE, ifnull(NUMERIC_SCALE,0), ifnull(NUMERIC_PRECISION,0), ifnull(CHARACTER_MAXIMUM_LENGTH,0),  default FROM qsys2.sysparms WHERE SPECIFIC_NAME='%s' and   SPECIFIC_SCHEMA ='%s' ORDER BY ORDINAL_POSITION", strings.ToUpper(sp.SpecificName), strings.ToUpper(sp.SpecificLib))

	sp.Parameters = make([]*StoredProcParamter, 0)
	conn, err := s.GetConnection()

	if err != nil {

		return err
	}

	rows, err := conn.Query(sql)
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
			log.Println("GetSPParameter ", err.Error())
		}

		sp.Parameters = append(sp.Parameters, spParamter)

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
			u.ID = uuid.NewString()
		}
		u.Name = strings.ToUpper(strings.TrimSpace(u.Name))
		u.Lib = strings.ToUpper(strings.TrimSpace(u.Lib))

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
func (m *StoredProcModel) DuplicateName(name string) bool {
	exists := false
	for _, savedQuery := range m.List() {

		if strings.EqualFold(savedQuery.Name, name) {
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
