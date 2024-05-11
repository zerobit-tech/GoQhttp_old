package storedProc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/zerobit-tech/GoQhttp/internal/inbuiltparam"
	"github.com/zerobit-tech/GoQhttp/internal/validator"
	"github.com/zerobit-tech/GoQhttp/logger"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
)

// type LogByType struct {
// 	Text string `json:"-" db:"-" form:"-"`
// 	Type string `json:"-" db:"-" form:"-"`
// }

type StoredProcResponse struct {
	ReferenceId string
	Status      int
	Message     string
	Data        map[string]any
	LogData     []*logger.LogEvent `json:"-" db:"-" form:"-"`
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

	HtmlTemplate string `json:"htmltemplate" db:"htmltemplate" form:"htmltemplate"`
	Namespace    string `json:"namespace" db:"namespace" form:"namespace"`

	IsSpecial     bool `json:"isspecial" db:"isspecial" form:"-"`
	MaxlogEntries int  `json:"maxlogentries" db:"maxlogentries" form:"maxlogentries"`
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------

type PreparedCallStatements struct {
	ResponseFormat         map[string]any
	InOutParams            []any // to send values to SP call
	InOutParamVariables    map[string]*any
	InOutParamMapToSPParam map[string]*StoredProcParamter
	FinalCallStatement     string
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) LogImage() string {
	imageMap := make(map[string]any)
	imageMap["EndPointName"] = s.EndPointName
	imageMap["HttpMethod"] = s.HttpMethod
	imageMap["Name"] = s.Name
	imageMap["Lib"] = s.Lib

	imageMap["SpecificName"] = s.SpecificName
	imageMap["SpecificLib"] = s.SpecificLib
	imageMap["UseSpecificName"] = s.UseSpecificName
	imageMap["DefaultServerId"] = s.DefaultServer.ID

	imageMap["AllowWithoutAuth"] = s.AllowWithoutAuth
	imageMap["Namespace"] = s.Namespace

	j, err := json.MarshalIndent(imageMap, " ", " ")
	if err == nil {
		return string(j)
	}

	return err.Error()
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) Slug() string {
	return slug.Make(s.Namespace + "_" + s.EndPointName + "_" + s.HttpMethod)

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) AvailableParamterPostions() []string {

	a := []string{"QUERY", "PATH"}

	if s.HttpMethod != "GET" && s.HttpMethod != "DELETE" {
		a = append(a, "BODY")
	}
	return a
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) AssignAliasForPathPlacement() {
	pathCounter := 0 // after api and endpoint name
	for _, p1 := range s.Parameters {
		if p1.Placement == "PATH" {
			p1.Alias = fmt.Sprintf("*PATH_%d", pathCounter)
			pathCounter += 1
		} else {
			if strings.HasPrefix(p1.Alias, "*PATH_") {
				p1.Alias = ""
			}
		}
	}
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) ValidateAlias() error {

	for _, p1 := range s.Parameters {
		for _, p2 := range s.Parameters {
			if p1.Name != p2.Name && p1.GetNameToUse(false) == p2.GetNameToUse(false) {
				return fmt.Errorf("Conflict between %s and %s.", p1.Name, p2.Name)
			}

		}

	}
	return nil
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) IsAllowedForServer(serverID string) bool {
	if serverID == "" {
		return false
	}

	for _, rcd := range s.AllowedOnServers {
		if serverID == rcd.ID {
			return true
		}
	}

	return false

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) AddAllowedServer(serverID, serverName string) {
	alreadyAssigned := false

	for _, rcd := range s.AllowedOnServers {
		if serverID == rcd.ID {
			alreadyAssigned = true
			rcd.Name = serverName
		}
	}

	if !alreadyAssigned {
		rcd := &ServerRecord{ID: serverID, Name: serverName}
		s.AllowedOnServers = append(s.AllowedOnServers, rcd)
	}

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) DeleteAllowedServer(serverID string) {

	a := make([]*ServerRecord, 0)

	for _, rcd := range s.AllowedOnServers {
		if serverID != rcd.ID {
			a = append(a, rcd)
		}
	}

	s.AllowedOnServers = a

}

// ------------------------------------------------------------
// BuildMockUrl(s)
// ------------------------------------------------------------
func (s *StoredProc) BuildMockUrlXX() {

	queryParamString := ""
	inputPayload := make(map[string]string)
	s.InputPayload = ""

outerloop:
	for _, p := range s.Parameters {
		if p.Mode == "OUT" {
			continue outerloop
		}
		nameToUse := p.GetNameToUse(false)
		// dont display inbuilt param
		for _, ibp := range inbuiltparam.InbuiltParams {
			if strings.EqualFold(ibp, nameToUse) {
				continue outerloop
			}
		}

		inputPayload[nameToUse] = fmt.Sprintf("{%s}", p.Datatype)

		if queryParamString == "" {
			queryParamString = fmt.Sprintf("?%s={%s}", nameToUse, p.Datatype)
		} else {
			queryParamString = queryParamString + fmt.Sprintf("&%s={%s}", nameToUse, p.Datatype)
		}

	}

	if s.HttpMethod != "GET" && s.HttpMethod != "DELETE" {

		jsonPayload, err := json.MarshalIndent(inputPayload, "", "  ")
		if err == nil {
			s.InputPayload = string(jsonPayload)
		}
		queryParamString = ""
	}

	s.MockUrl = fmt.Sprintf("api/%s%s", s.EndPointName, queryParamString)
	s.MockUrlWithoutAuth = fmt.Sprintf("uapi/%s%s", s.EndPointName, queryParamString)
}

// ------------------------------------------------------------
// BuildMockUrl(s)
// ------------------------------------------------------------
func (s *StoredProc) BuildMockUrl() {

	if strings.TrimSpace(s.Namespace) == "" {
		s.Namespace = "V1"
	}

	switch s.HttpMethod {
	case "GET", "DELETE":
		s.BuildMockUrlGET()
	default:
		s.BuildMockUrlPost()
	}
}

// ------------------------------------------------------------
// BuildMockUrl(s)
// ------------------------------------------------------------
func (s *StoredProc) BuildMockUrlGET() {

	queryParamString := ""
	s.InputPayload = ""
	pathParamString := ""
outerloop:
	for _, p := range s.Parameters {

		nameToUse := p.GetNameToUse(false)
		if p.Mode == "OUT" {
			continue outerloop
		}

		// dont display inbuilt param
		for _, ibp := range inbuiltparam.InbuiltParams {
			if strings.EqualFold(ibp, nameToUse) {
				continue outerloop
			}
		}

		switch p.Placement {
		case "PATH":
			pathParamString = pathParamString + fmt.Sprintf("/{%s__%s}", nameToUse, p.Datatype)

		default:
			if queryParamString == "" {
				queryParamString = fmt.Sprintf("?%s={%s}", nameToUse, p.Datatype)
			} else {
				queryParamString = queryParamString + fmt.Sprintf("&%s={%s}", nameToUse, p.Datatype)
			}
		}

	}

	s.MockUrl = fmt.Sprintf("api/%s/%s%s%s", s.Namespace, s.EndPointName, pathParamString, queryParamString)
	s.MockUrlWithoutAuth = fmt.Sprintf("uapi/%s/%s%s%s", s.Namespace, s.EndPointName, pathParamString, queryParamString)
}

// ------------------------------------------------------------
// BuildMockUrl(s)
// ------------------------------------------------------------
func (s *StoredProc) BuildMockUrlPost() {

	queryParamString := ""
	inputPayload := make(map[string]string)
	s.InputPayload = ""
	pathParamString := ""
outerloop:
	for _, p := range s.Parameters {
		if p.Mode == "OUT" {
			continue outerloop
		}

		nameToUse := p.GetNameToUse(false)

		// dont display inbuilt param
		for _, ibp := range inbuiltparam.InbuiltParams {
			if strings.EqualFold(ibp, nameToUse) {
				continue outerloop
			}
		}

		switch p.Placement {
		case "PATH":
			pathParamString = pathParamString + fmt.Sprintf("/{%s__%s}", nameToUse, p.Datatype)
		case "QUERY":
			if queryParamString == "" {
				queryParamString = fmt.Sprintf("?%s={%s}", nameToUse, p.Datatype)
			} else {
				queryParamString = queryParamString + fmt.Sprintf("&%s={%s}", nameToUse, p.Datatype)
			}
		default:
			inputPayload[nameToUse] = fmt.Sprintf("{%s}", p.Datatype)
		}
	}

	jsonPayload, err := json.MarshalIndent(inputPayload, "", "  ")
	if err == nil {
		s.InputPayload = string(jsonPayload)
	}

	s.MockUrl = fmt.Sprintf("api/%s/%s%s%s", s.Namespace, s.EndPointName, pathParamString, queryParamString)
	s.MockUrlWithoutAuth = fmt.Sprintf("uapi/%s/%s%s%s", s.Namespace, s.EndPointName, pathParamString, queryParamString)
}

// ------------------------------------------------------------
// set name space value
// ------------------------------------------------------------
func (s *StoredProc) SetNameSpace() {
	s.Namespace = strings.TrimSpace(s.Namespace)
	if strings.TrimSpace(s.Namespace) == "" {
		s.Namespace = "v1"
	}

	s.Namespace = stringutils.RemoveSpecialChars(stringutils.RemoveMultipleSpaces(s.Namespace))
	s.Namespace = strings.ToLower(strings.TrimSpace(s.Namespace))
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) GetNamespace() string {
	s.SetNameSpace()
	return s.Namespace
}
