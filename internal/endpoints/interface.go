package endpoints

import "html/template"

type Endpoint interface {
	EPType() string // rpg or sql
	EPNameSpace() string
	EPName() string
	EPMethod() string
	EPServerId() string
	EPHander() string
	EPDetailUrl() template.HTML
}
