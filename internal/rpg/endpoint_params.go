package rpg

import (
	"strings"

	"github.com/zerobit-tech/GoQhttp/internal/validator"
)

// -----------------------------------------------------
//
// -----------------------------------------------------
type ProgramParams struct {
	Seq                 uint
	Dim                 uint
	InOutType           string
	FieldID             string
	Placement           string
	NameToUse           string
	Param               *Param                     `json:"-" db:"-" form:"-"`
	validator.Validator `json:"-" db:"-" form:"-"` // this contains the fielderror

}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (p *ProgramParams) GetNameToUse() string {

	return strings.ToUpper(p.NameToUse)
}
