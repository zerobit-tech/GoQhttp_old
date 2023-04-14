package xmlutils

import (
	"encoding/xml"
	"io"
	"strings"
)

func IsValid(input string) bool {

	err := xml.Unmarshal([]byte(input), new(interface{}))
	if err != nil {
		return false
	}

	decoder := xml.NewDecoder(strings.NewReader(input))

	val := new(interface{})
	for {
		err := decoder.Decode(val)
		if err != nil {
			return err == io.EOF
		}
	}
}
