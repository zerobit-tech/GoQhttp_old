package rpg

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/onlysumitg/GoQhttp/internal/inbuiltparam"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/utils/stringutils"
)

type RpgEndPoint struct {
	ID                 string   `json:"id" db:"id" form:"id"`
	EndPointName       string   `json:"endpointname" db:"endpointname" form:"endpointname"`
	HttpMethod         string   `json:"httpmethod" db:"httpmethod" form:"httpmethod"`
	RpgProgram         string   `json:"rpgpgmid" db:"rpgpgmid" form:"rpgpgmid"`
	DefaultServerId    string   `json:"serverid" db:"serverid" form:"serverid"`
	AllowedOnServers   []string `json:"allowedonservers" db:"allowedonservers" form:"allowedonservers"`
	MockUrl            string   `json:"mockurl" db:"mockurl" form:"-"`
	MockUrlWithoutAuth string   `json:"mockurlnoa" db:"mockurlnoa" form:"-"`
	AllowWithoutAuth   bool     `json:"awoauth" db:"awoauth" form:"awoauth"`
	HtmlTemplate       string   `json:"htmltemplate" db:"htmltemplate" form:"htmltemplate"`
	Namespace          string   `json:"namespace" db:"namespace" form:"namespace"`
	InputPayload       string   `json:"inputpayload" db:"inputpayload" form:"inputpayload"`
	ResponseFormat     string   `json:"responseformat" db:"responseformat" form:"-"`
	Promotionsql       string   `json:"promotionsql" db:"promotionsql" form:"-"`

	RpgPgm *Program `json:"-" db:"-" form:"-"`

	validator.Validator `json:"-" db:"-" form:"-"` // this contains the fielderror
}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (s *RpgEndPoint) Refresh() {

	s.SetNameSpace()
	s.BuildMockUrl()
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) Slug() string {
	return slug.Make(s.Namespace + "_" + s.EndPointName + "_" + s.HttpMethod)

}

// ------------------------------------------------------------
// set name space value
// ------------------------------------------------------------
func (s *RpgEndPoint) SetNameSpace() {
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
func (s *RpgEndPoint) GetNamespace() string {
	s.SetNameSpace()
	return s.Namespace
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) AddAllowedServer(serverID string) {
	alreadyAssigned := false

	for _, rcd := range s.AllowedOnServers {
		if serverID == rcd {
			alreadyAssigned = true
		}
	}

	if !alreadyAssigned {
		rcd := serverID
		s.AllowedOnServers = append(s.AllowedOnServers, rcd)
	}

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) DeleteAllowedServer(serverID string) {

	a := make([]string, 0)

	for _, rcd := range s.AllowedOnServers {
		if serverID != rcd {
			a = append(a, rcd)
		}
	}

	s.AllowedOnServers = a

}

// ------------------------------------------------------------
// BuildMockUrl(s)
// ------------------------------------------------------------
func (s *RpgEndPoint) BuildMockUrl() {

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
func (s *RpgEndPoint) BuildMockUrlGET() {

	queryParamString := ""
	s.InputPayload = ""
	pathParamString := ""
outerloop:
	for _, p := range s.RpgPgm.Parameters {

		nameToUse := p.getNameToUse()
		if p.InOutType == "OUT" {
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
			pathParamString = pathParamString + fmt.Sprintf("/{%s__%s}", nameToUse, p.Param.DataType)

		default:
			if queryParamString == "" {
				queryParamString = fmt.Sprintf("?%s={%s}", nameToUse, p.Param.DataType)
			} else {
				queryParamString = queryParamString + fmt.Sprintf("&%s={%s}", nameToUse, p.Param.DataType)
			}
		}

	}

	s.MockUrl = fmt.Sprintf("api/%s/%s%s%s", s.Namespace, s.EndPointName, pathParamString, queryParamString)
	s.MockUrlWithoutAuth = fmt.Sprintf("uapi/%s/%s%s%s", s.Namespace, s.EndPointName, pathParamString, queryParamString)
}

// ------------------------------------------------------------
// BuildMockUrl(s)
// ------------------------------------------------------------
func (s *RpgEndPoint) BuildMockUrlPost() {

	queryParamString := ""
	inputPayload := make(map[string]string)
	s.InputPayload = ""
	pathParamString := ""
outerloop:
	for _, p := range s.RpgPgm.Parameters {

		if p.Param == nil {
			continue
		}

		if p.InOutType == "OUT" {
			continue outerloop
		}

		nameToUse := p.getNameToUse()

		// dont display inbuilt param
		for _, ibp := range inbuiltparam.InbuiltParams {
			if strings.EqualFold(ibp, nameToUse) {
				continue outerloop
			}
		}

		switch p.Placement {
		case "PATH":
			pathParamString = pathParamString + fmt.Sprintf("/{%s__%s}", nameToUse, p.Param.DataType)
		case "QUERY":
			if queryParamString == "" {
				queryParamString = fmt.Sprintf("?%s={%s}", nameToUse, p.Param.DataType)
			} else {
				queryParamString = queryParamString + fmt.Sprintf("&%s={%s}", nameToUse, p.Param.DataType)
			}
		default:
			inputPayload[nameToUse] = fmt.Sprintf("{%s}", p.Param.DataType)
		}
	}

	jsonPayload, err := json.MarshalIndent(inputPayload, "", "  ")
	if err == nil {
		s.InputPayload = string(jsonPayload)
	}

	s.MockUrl = fmt.Sprintf("api/%s/%s%s%s", s.Namespace, s.EndPointName, pathParamString, queryParamString)
	s.MockUrlWithoutAuth = fmt.Sprintf("uapi/%s/%s%s%s", s.Namespace, s.EndPointName, pathParamString, queryParamString)
}
