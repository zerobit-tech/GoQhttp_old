package rpg

import (
	"strings"

	"github.com/gosimple/slug"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/utils/stringutils"
)

type RpgEndPoint struct {
	ID                  string                     `json:"id" db:"id" form:"id"`
	EndPointName        string                     `json:"endpointname" db:"endpointname" form:"endpointname"`
	HttpMethod          string                     `json:"httpmethod" db:"httpmethod" form:"httpmethod"`
	RpgProgram          string                     `json:"rpgpgmid" db:"rpgpgmid" form:"rpgpgmid"`
	DefaultServerId     string                     `json:"-" db:"serverid" form:"serverid"`
	AllowedOnServers    []string                   `json:"allowedonservers" db:"allowedonservers" form:"allowedonservers"`
	MockUrl             string                     `json:"mockurl" db:"mockurl" form:"-"`
	MockUrlWithoutAuth  string                     `json:"mockurlnoa" db:"mockurlnoa" form:"-"`
	AllowWithoutAuth    bool                       `json:"awoauth" db:"awoauth" form:"awoauth"`
	HtmlTemplate        string                     `json:"htmltemplate" db:"htmltemplate" form:"htmltemplate"`
	Namespace           string                     `json:"namespace" db:"namespace" form:"namespace"`
	InputPayload        string                     `json:"inputpayload" db:"inputpayload" form:"inputpayload"`
	validator.Validator `json:"-" db:"-" form:"-"` // this contains the fielderror
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
