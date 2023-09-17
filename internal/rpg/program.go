package rpg

import (
	"fmt"
	"strings"

	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
)

type ProgramParams struct {
	Seq       uint    
	InOutType string  
	FieldID   string  
}

type Program struct {
	ID                  string `json:"id" db:"id" form:"id"`
	Name                string `json:"name" db:"name" form:"name"`
	Lib                 string `json:"lib" db:"lib" form:"lib"`
	Parameters          []*ProgramParams
	validator.Validator `json:"-" db:"-" form:"-"` // this contains the fielderror
}

func (p *Program) Init() {

	p.Parameters = make([]*ProgramParams, 0, 20)

	for i := 0; i < 20; i++ {

		pp := &ProgramParams{}
		p.Parameters = append(p.Parameters, pp)

	}

}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (p *Program) ToXML(inparams map[string]xmlutils.ValueDatatype) string {
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

	xmlString := fmt.Sprintf(`<?xml version="1.0" ?><xmlservice><pgm error="off" lib="%s" name="%s" var="%s"> %s</pgm></xmlservice>`, p.Lib, p.Name, p.Name, strings.Join(p.ParamStrings(inparams), "\n"))

	return xmlString
}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (p *Program) ParamStrings(inparams map[string]xmlutils.ValueDatatype) []string {
	parms := make([]string, 0)

	// for _, pr := range p.Parameters {

	// 	valX, found := inparams[strings.ToUpper(pr.Name)]
	// 	valS := ""
	// 	if found {

	// 		valS = stringutils.AsString(valX.Value)

	// 	}
	// 	parms = append(parms, pr.ToXml(valS))
	// }

	return parms
}
