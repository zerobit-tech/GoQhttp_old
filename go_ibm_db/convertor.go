package go_ibm_db

import (
	"reflect"
	"strings"
)

type Col struct {
	Index int `json:"-"`
	Name  string
	Type  reflect.Type `json:"-"`
	Value any
}

type RSRows []Col

func (c *Col) AssignValueByType() {

	switch c.Type.Kind() {
	case reflect.String:
		if c.Value == nil {
			c.Value = ""
		} else {
			c.Value = strings.TrimSpace(asString(c.Value))
		}
	}
}

func (r RSRows) ToMap() map[string]any {
	m := make(map[string]any)

	for _, c := range r {
		c.AssignValueByType()
		m[c.Name] = c.Value
	}

	return m
}
