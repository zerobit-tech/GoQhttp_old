package endpoints

import "html/template"

type Endpoint interface {
	EPType() string // rpg or sql
	EPID() string
	EPMaxLogEntries() int

	EPNameSpace() string
	EPName() string
	EPMethod() string
	EPServerId() string
	EPHander() string
	EPDetailUrl() template.HTML
}
