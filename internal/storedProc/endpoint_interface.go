package storedProc

import (
	"fmt"
	"html/template"
)

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) EPID() string {
	return s.ID
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) EPMaxLogEntries() int {
	return s.MaxlogEntries
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) EPType() string {
	return "SP"
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) EPNameSpace() string {
	return s.GetNamespace()
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) EPName() string {
	return s.EndPointName
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) EPMethod() string {
	return s.HttpMethod
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) EPServerId() string {
	return s.DefaultServer.ID
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) EPHander() string {
	return fmt.Sprintf("%s/%s", s.SpecificLib, s.SpecificName)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *StoredProc) EPDetailUrl() template.HTML {

	return template.HTML(fmt.Sprintf("<a  hx-push-url='true'  class='btn btn-ghost-info' href='/sp/%s'>View</a>", s.ID))
}
