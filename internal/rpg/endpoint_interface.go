package rpg

import (
	"fmt"
	"html/template"
)

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) EPID() string {
	return s.ID
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) EPMaxLogEntries() int {
	return s.MaxlogEntries
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) EPType() string {
	return "PGM"
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) EPNameSpace() string {
	return s.GetNamespace()
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) EPName() string {
	return s.EndPointName
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) EPMethod() string {
	return s.HttpMethod
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) EPServerId() string {
	return s.DefaultServerId
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) EPHander() string {
	return fmt.Sprintf("%s/%s", s.Name, s.Lib)
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *RpgEndPoint) EPDetailUrl() template.HTML {

	return template.HTML(fmt.Sprintf("<a  hx-push-url='true'  class='btn btn-ghost-info' href='/pgmendpoints/%s'>View</a>", s.ID))
}
