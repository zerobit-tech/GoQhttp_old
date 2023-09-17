package rpg

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/onlysumitg/GoQhttp/internal/validator"
)

type Param struct {
	ID string `json:"id" db:"id" form:"id"`

	Name            string `json:"name" db:"name" form:"name"`
	DataType        string `json:"datatype" db:"datatype" form:"datatype"`
	Length          uint   `json:"length" db:"length" form:"length"`
	DecimalPostions uint   `json:"decimalpostions" db:"decimalpostions" form:"decimalpostions"`
	IsVarying       bool   `json:"isvarying" db:"isvarying" form:"isvarying"`

	IsDs     bool     `json:"isds" db:"isds" form:"isds"`
	DsFields []string `json:"dsfields" db:"dsfields" form:"dsfields"`
	DsDim    uint     `json:"dsdim" db:"dsdim" form:"dsdim"`

	validator.Validator `json:"-" db:"-" form:"-"` // this contains the fielderror

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *Param) ToString() string {

	if p.IsDs {
		return fmt.Sprintf("DS: %s Dim %d ", p.Name, p.DsDim)

	} else {
		baseString := fmt.Sprintf("%s %s", p.Name, p.DataType)

		if DataTypeNeedDecimalValue(p.DataType) {
			baseString = fmt.Sprintf("%s (%d : %d)", baseString, p.Length, p.DecimalPostions)
		} else if DataTypeNeedLength(p.DataType) {
			baseString = fmt.Sprintf("%s (%d) ", baseString, p.Length)

		}

		return baseString

	}

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (s *Param) LogImage() string {
	imageMap := make(map[string]any)
	imageMap["Name"] = s.Name
	imageMap["DataType"] = s.DataType
	imageMap["Length"] = s.Length

	imageMap["DecimalPostions"] = s.DecimalPostions

	imageMap["IsVarying"] = s.IsVarying

	j, err := json.MarshalIndent(imageMap, " ", " ")
	if err == nil {
		return string(j)
	}

	return err.Error()
}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (p *Param) ToXml(value string) string {
	/*
					  <parm io="both" var="p4">
				            <data type="12p2" var="INDEC2">1234567890.12</data>
				        </parm>


		for DS
		        <parm io="both" var="p5">
		            <ds var="INDS1">
		                <data type="1a" var="DSCHARA">E</data>
		                <data type="1a" var="DSCHARB">F</data>
		                <data type="7p4" var="DSDEC1">333.3330</data>
		                <data type="12p2" var="DSDEC2">4444444444.44</data>
		            </ds>
		        </parm>
	*/
	pramString := ""
	pname := p.Name
	dsname := fmt.Sprintf("DS%s", p.Name)

	//         {'io':'in|out|both|omit'} : XMLSERVICE param type {'io':both'}.
	//                by='val|ref'

	if p.IsDs {
		pramString = fmt.Sprintf("<parm io=\"both\" var=\"%s\"><ds var=\"%s\"> %s  </ds></parm>", pname, dsname, strings.Join(p.ToDataXmlList(), "\n"))
	} else {

		pramString = fmt.Sprintf("<parm io=\"both\" var=\"%s\"> %s </parm>", pname, p.ToDataXml(value))
	}
	return pramString
}

// -----------------------------------------------------
//
//	DS fields
//
// -----------------------------------------------------
func (p *Param) ToDataXmlList() []string {
	/*
		   <data type="12p2" var="INDEC2"><![CDATA[1234567890.12].12</data>
		     <data type="12p2" var="INDEC2"><![CDATA[1234567890.12].12</data>
			   <data type="12p2" var="INDEC2"><![CDATA[1234567890.12].12</data>
	*/
	items := make([]string, 0)

	// for _, x := range p.DsFields {

	// 	items = append(items, x.ToDataXml("TODO"))
	// }

	return items

}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (p *Param) ToDataXml(value string) string {
	/*
	   <data type="12p2" var="INDEC2"><![CDATA[1234567890.12]</data>
	*/

	dateType := p.GetDataType()
	dname := fmt.Sprintf("T%s", p.Name)
	//value := p.Value

	pramString := fmt.Sprintf("  <data type=\"%s\" var=\"%s\"><![CDATA[%s]]></data>  ", dateType, dname, value)
	return pramString
}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (p *Param) GetDataType() string {
	dataType, found := DataTypeMap[p.DataType]
	if found {
		_, foundLength := dataTypeWithLength[p.DataType]
		_, foundDecimal := dataTypeWithDecimal[p.DataType]

		switch {
		case foundLength && foundDecimal:
			dataType = fmt.Sprintf(dataType, p.Length, p.DecimalPostions)
		case foundLength:
			dataType = fmt.Sprintf(dataType, p.Length)

		}
	}

	return dataType
}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (p *Param) IsValid() error {
	_, found := DataTypeMap[p.DataType]
	if !found {
		p.CheckField(false, "datatype", "Invalid data type")

		return errors.New("Invalid data type")
	}

	_, found = dataTypeWithLength[p.DataType]
	if found && p.Length <= 0 {
		p.CheckField(false, "length", "Length is required")

		return errors.New("Length is required")
	}
	return nil

}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (p *Param) HasValidValue() error {

	return nil
}
