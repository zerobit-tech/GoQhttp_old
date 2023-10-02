package rpg

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/onlysumitg/GoQhttp/internal/inbuiltparam"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/utils/stringutils"
	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
)

// -----------------------------------------------------
//
// -----------------------------------------------------

type RpgEndPoint struct {
	ID                 string   `json:"id" db:"id" form:"id"`
	EndPointName       string   `json:"endpointname" db:"endpointname" form:"endpointname"`
	HttpMethod         string   `json:"httpmethod" db:"httpmethod" form:"httpmethod"`
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

	//	RpgProgram         string   `json:"rpgpgmid" db:"rpgpgmid" form:"rpgpgmid"`
	//	RpgPgm *Program `json:"-" db:"-" form:"-"`

	Name       string `json:"name" db:"name" form:"name"`
	Lib        string `json:"lib" db:"lib" form:"lib"`
	Parameters []*ProgramParams

	validator.Validator `json:"-" db:"-" form:"-"` // this contains the fielderror
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *RpgEndPoint) AssignParamObjects(rpgParamModel *RpgParamModel) {
	for _, f := range p.Parameters {

		param, err := rpgParamModel.Get(f.FieldID)
		if err == nil {
			f.Param = param
		}

	}
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *RpgEndPoint) AssignParamNames() {

	for _, f := range p.Parameters {
		if strings.TrimSpace(f.NameToUse) == "" && f.Param != nil {
			f.NameToUse = strings.ToUpper(f.Param.Name)
		}
	}

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *RpgEndPoint) FilterOutInvalidParams() {

	fields := make([]*ProgramParams, 0)
	for _, f := range p.Parameters {
		field := f
		f.NameToUse = strings.TrimSpace(strings.ToUpper(f.NameToUse))

		if f.NameToUse != "" && f.FieldID != "" {
			fields = append(fields, field)
		}
	}

	p.Parameters = fields

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *RpgEndPoint) ValidateParams() bool {
	anyError := false

	anyError = p.checkNameAndParamAssigned()
	if !anyError {
		anyError = p.checkDuplicateFieldName()
	}

	return anyError

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *RpgEndPoint) checkNameAndParamAssigned() bool {
	anyError := false

	for _, f := range p.Parameters {
		if f.NameToUse == "" && f.FieldID != "" {
			f.AddFieldError("name", "Cannot be blank")
			anyError = true
		}

		if f.NameToUse != "" && f.FieldID == "" {
			f.AddFieldError("name", "Name without the field")
			anyError = true
		}
	}

	return anyError

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *RpgEndPoint) checkDuplicateFieldName() bool {
	anyError := false

	nameMap := make(map[string]bool)
	for _, f := range p.Parameters {
		if strings.TrimSpace(f.NameToUse) == "" {
			continue
		}
		f.NameToUse = strings.ToUpper(f.NameToUse)
		_, found := nameMap[f.NameToUse]
		if found {
			f.AddFieldError("name", "Duplicate Name")
			anyError = true

		} else {
			nameMap[f.NameToUse] = true
		}
	}

	return anyError
}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (p *RpgEndPoint) Init() {

	// p.Parameters = make([]*ProgramParams, 0, 20)

	// for i := 0; i < 20; i++ {

	// 	pp := &ProgramParams{}
	// 	p.Parameters = append(p.Parameters, pp)

	// }

}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (s *RpgEndPoint) Refresh() {

	s.SetNameSpace()
	s.AssignAliasForPathPlacement()
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
	for _, p := range s.Parameters {

		if p.Param == nil {
			continue
		}

		nameToUse := p.getNameToUse()
		if p.InOutType == "out" {
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
func (s *RpgEndPoint) InputParamJson() string {

	outJsonString := ""
	inputPayload := make(map[string]any)

outerloop:
	for _, p := range s.Parameters {

		if p.Param == nil {
			continue
		}

		if p.InOutType == "out" {
			continue outerloop
		}

		switch p.Placement {
		case "PATH":
		case "QUERY":
		default:
			if p.Param.IsDs {

				x := p.Param.DsJson(p.Dim)
				if p.Dim > 1 {
					inputPayload[p.NameToUse] = x
				} else {
					inputPayload[p.NameToUse] = x[0]
				}

			} else {

				x := p.Param.NoNDsJson(p.Dim)
				if p.Dim > 1 {
					inputPayload[p.NameToUse] = x
				} else {
					inputPayload[p.NameToUse] = x[0]
				}

			}

		}
	}
	jsonPayload, err := json.MarshalIndent(inputPayload, "", "  ")
	if err == nil {
		outJsonString = string(jsonPayload)
	}

	return outJsonString

}

// ------------------------------------------------------------
// BuildMockUrl(s)
// ------------------------------------------------------------
func (s *RpgEndPoint) BuildMockUrlPost() {

	queryParamString := ""
	s.InputPayload = ""
	pathParamString := ""
outerloop:
	for _, p := range s.Parameters {

		if p.Param == nil {
			continue
		}

		if p.InOutType == "out" {
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

		}
	}

	s.InputPayload = s.InputParamJson()

	s.MockUrl = fmt.Sprintf("api/%s/%s%s%s", s.Namespace, s.EndPointName, pathParamString, queryParamString)
	s.MockUrlWithoutAuth = fmt.Sprintf("uapi/%s/%s%s%s", s.Namespace, s.EndPointName, pathParamString, queryParamString)
}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (p *RpgEndPoint) ToXML(inparams map[string]xmlutils.ValueDatatype) (string, error) {
	/*
		<?xml version="1.0" ?>
		<xmlservice>
			<pgm error="fast" lib="SUMITG1" name="QHTTPTEST1" var="QHTTPTEST1">
				<parm io="both" var="p1">
					<data type="5s2" var="I1">10</data>
				</parm>
				<parm io="both" var="p2">
					<data type="5s2" var="I2">20</data>
				</parm>
				<parm io="both" var="p3">
					<data type="5s2" var="SUM">0</data>
				</parm>

			</pgm>
		</xmlservice>
	*/

	xmlStrings, err := p.ParamStrings(inparams)
	if err != nil {
		return "", err
	}

	xmlString := fmt.Sprintf(`<?xml version="1.0" ?><xmlservice><pgm error="off" lib="%s" name="%s" var="%s"> %s</pgm></xmlservice>`, p.Lib, p.Name, p.Name, strings.Join(xmlStrings, "\n"))

	return xmlString, nil
}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (p *RpgEndPoint) ParamStrings(inparams map[string]xmlutils.ValueDatatype) ([]string, error) {
	parms := make([]string, 0)

	for _, pr := range p.Parameters {

		if pr.Param == nil {
			continue
		}

		xmlString, err := pr.Param.ToXml(pr.getNameToUse(), inparams, pr.InOutType, pr.Dim)
		if err != nil {
			return make([]string, 0), err
		}

		parms = append(parms, xmlString)
	}

	return parms, nil
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) AvailableParamterPostions() []string {

	a := []string{"QUERY", "PATH"}

	if s.HttpMethod != "GET" && s.HttpMethod != "DELETE" {
		a = append(a, "BODY")
	}
	return a
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) AssignAliasForPathPlacement() {
	pathCounter := 0 // after api and endpoint name
	for _, p1 := range s.Parameters {
		if p1.Placement == "PATH" {
			p1.NameToUse = fmt.Sprintf("*PATH_%d", pathCounter)
			pathCounter += 1
		} else {
			if strings.HasPrefix(p1.NameToUse, "*PATH_") {
				p1.NameToUse = strings.TrimLeft(p1.NameToUse, "*")
			}
		}
	}
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) IsAllowedForServer(serverID string) bool {
	if serverID == "" {
		return false
	}

	for _, rcd := range s.AllowedOnServers {
		if serverID == rcd {
			return true
		}
	}

	return false

}
