package storedProc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/onlysumitg/GoQhttp/internal/validator"
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
// BuildMockUrl(s)
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
// BuildMockUrl(s)
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
