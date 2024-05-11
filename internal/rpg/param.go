package rpg

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/zerobit-tech/GoQhttp/internal/validator"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
	"github.com/zerobit-tech/GoQhttp/utils/xmlutils"
)

// -----------------------------------------------------
//
// -----------------------------------------------------
type DSField struct {
	NameToUse           string
	ParamID             string
	Dim                 uint
	Param               *Param            `json:"-" db:"-" form:"-"`
	validator.Validator `db:"-" form:"-"` // this contains the fielderror

}

type Param struct {
	ID string `json:"id" db:"id" form:"id"`

	Name            string `json:"name" db:"name" form:"name"`
	DataType        string `json:"datatype" db:"datatype" form:"datatype"`
	Length          uint   `json:"length" db:"length" form:"length"`
	DecimalPostions uint   `json:"decimalpostions" db:"decimalpostions" form:"decimalpostions"`

	IsDs     bool `json:"isds" db:"isds" form:"isds"`
	DsFields []*DSField

	validator.Validator `json:"-" db:"-" form:"-"` // this contains the fielderror

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *Param) DsHasField(fieldId string) bool {
	if p.IsDs {
		for _, f := range p.DsFields {
			if f != nil && f.ParamID == fieldId {
				return true
			}
		}
	}

	return false
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *Param) AssignDSFieldNames() {
	if p.IsDs {
		for _, f := range p.DsFields {
			if strings.TrimSpace(f.NameToUse) == "" && f.Param != nil {
				f.NameToUse = strings.ToUpper(f.Param.Name)
			}
		}
	}
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *Param) FilterOutInvalidParams() {
	if p.IsDs {
		fields := make([]*DSField, 0)
		for _, f := range p.DsFields {
			field := f
			f.NameToUse = strings.TrimSpace(strings.ToUpper(f.NameToUse))

			if f.NameToUse != "" && f.ParamID != "" {
				fields = append(fields, field)
			}
		}

		p.DsFields = fields
	}

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *Param) ValidateFields() bool {
	anyError := false
	if p.IsDs {
		anyError = p.checkNameAndParamAssigned()
		if !anyError {
			anyError = p.checkDuplicateFieldName()
		}
	}
	return anyError

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *Param) checkNameAndParamAssigned() bool {
	anyError := false
	if p.IsDs {
		for _, f := range p.DsFields {
			if f.NameToUse == "" && f.ParamID != "" {
				f.AddFieldError("name", "Cannot be blank")
				anyError = true
			}

			if f.NameToUse != "" && f.ParamID == "" {
				f.AddFieldError("name", "Name without the field")
				anyError = true
			}
		}

	}
	return anyError

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *Param) checkDuplicateFieldName() bool {
	anyError := false
	if p.IsDs {

		nameMap := make(map[string]bool)
		for _, f := range p.DsFields {
			if strings.TrimSpace(f.NameToUse) == "" {
				continue
			}
			f.NameToUse = strings.ToUpper(f.NameToUse)
			_, found := nameMap[f.NameToUse]
			if found {
				f.AddFieldError("name", "Duplicate Name")
				anyError = true

			} else {
				nameMap[f.NameToUse] = true
			}
		}
	}

	return anyError
}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *Param) Init() {

	// p.DsFields = make([]*DSField, 20)

	// for i := 0; i < 20; i++ {
	// 	p.DsFields[i] = &DSField{}
	// }

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *Param) NoNDsJson(dim uint) []string {
	if dim < 1 {
		dim = 1
	}
	inputPayloadList := make([]string, dim)
	if p.IsDs {
		return inputPayloadList
	}

	for i := 0; i < int(dim); i++ {
		inputPayloadList[i] = fmt.Sprintf("{%s}", p.DataType)
	}

	return inputPayloadList

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *Param) DsJson(dim uint) []map[string]any {
	if dim < 1 {
		dim = 1
	}

	inputPayloadList := make([]map[string]any, dim)
	if !p.IsDs {
		return inputPayloadList
	}

	for i := 0; i < int(dim); i++ {

		inputPayload := make(map[string]any)
		//inputPayloadList[i] =
		for _, f := range p.DsFields {
			if f.Param == nil {
				continue
			}

			if f.Param.IsDs {
				listItems := f.Param.DsJson(f.Dim)

				if f.Dim > 1 {
					inputPayload[f.NameToUse] = listItems

				} else {
					inputPayload[f.NameToUse] = listItems[0]
				}

			} else {

				//these can also repeatf

				listItems := f.Param.NoNDsJson(f.Dim)
				if f.Dim > 1 {
					inputPayload[f.NameToUse] = listItems

				} else {
					inputPayload[f.NameToUse] = listItems[0]
				}

			}

		}
		inputPayloadList[i] = inputPayload

	}

	return inputPayloadList

}

// ------------------------------------------------------------
//
// ------------------------------------------------------------
func (p *Param) ToString() string {

	if p.IsDs {
		return fmt.Sprintf("DS: %s", p.Name)

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

	j, err := json.MarshalIndent(imageMap, " ", " ")
	if err == nil {
		return string(j)
	}

	return err.Error()
}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (p *Param) ToXml(nameToUse string, inparams map[string]xmlutils.ValueDatatype, usage string, dim uint) (string, error) {
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
	pname := nameToUse
	//dsname := nameToUse //fmt.Sprintf("DS%s", nameToUse)

	//         {'io':'in|out|both|omit'} : XMLSERVICE param type {'io':both'}.
	//                by='val|ref'

	if p.IsDs {
		xmlParamString, err := p.dsXML(nameToUse, nameToUse, inparams, dim)
		if err != nil {
			return "", err
		}
		pramString = fmt.Sprintf("<parm io=\"%s\" var=\"%s\">   %s  </parm>", usage, pname, strings.Join(xmlParamString, "\n"))
	} else {
		paramString, err := p.singleParameterXMLXX(nameToUse, nameToUse, inparams, dim)
		if err != nil {
			return "", err
		}
		pramString = fmt.Sprintf("<parm io=\"%s\" var=\"%s\"> %s </parm>", usage, pname, paramString)
	}
	return pramString, nil
}

// -----------------------------------------------------
//
//	DS fields
//
// -----------------------------------------------------
func (p *Param) dsXML(key string, name string, values map[string]xmlutils.ValueDatatype, dim uint) ([]string, error) {
	/*

				// single ds =======================================================================
				<ds var="INDS1">
					<data type="12p2" var="INDEC2"><![CDATA[1234567890.12].12</data>
				    <data type="12p2" var="INDEC2"><![CDATA[1234567890.12].12</data>
					<data type="12p2" var="INDEC2"><![CDATA[1234567890.12].12</data>
			 	</ds>


			  // multi dim ds ======================================================================
			  <parm >
			  	<ds var="a1">1</ds>
				<ds var="a1">2</ds>
				<ds var="a1">3</ds>
			  </parm>


				// nested ds =======================================================================
				<ds var="INDS1">
		                <data type="1a" var="DSCHARA"> <![CDATA[a]]></data>
		                <data type="1a" var="DSCHARB"><![CDATA[b]]></data>
		                <data type="7p4" var="DSDEC1"> <![CDATA[32.1234]]></data>
		                <data type="12p2" var="DSDEC2"><![CDATA[33.33]]></data>

		                <ds var="NESTEDDS">
		                    <data type="1a" var="NESTED_DSCHARA"> <![CDATA[a]]></data>    <![CDATA[32.1234]]></data>
		                    <data type="7p4" var="NESTED_DSDEC1">
		                </ds>

						<ds var="NESTEDDS">
		                    <data type="1a" var="NESTED_DSCHARA"> <![CDATA[a]]></data>    <![CDATA[32.1234]]></data>
		                    <data type="7p4" var="NESTED_DSDEC1">
		                </ds>
		        </ds>
	*/

	items := make([]string, 0)
	if dim > 1 {

		/*
			<ds var="a1">1</ds>
			<ds var="a1">2</ds>
			<ds var="a1">3</ds>
		*/
		for i := 0; i < int(dim); i++ {
			key := fmt.Sprintf("%s[%d]", strings.ToUpper(key), i)

			xmlParamString, err := p.dsXML(key, name, values, 0)
			if err != nil {
				return make([]string, 0), err
			}

			strToUse := strings.Join(xmlParamString, "\n")
			items = append(items, strToUse)

		}

	} else {

		items = append(items, fmt.Sprintf("<ds var=\"%s\">", name))

		for _, x := range p.DsFields {
			if x.Param == nil {
				continue
			}

			if x.Param.IsDs {
				nameForNestedDS := x.NameToUse
				keyForNestedDS := fmt.Sprintf("%s.%s", key, x.NameToUse)
				xmlParamString, err := x.Param.dsXML(keyForNestedDS, nameForNestedDS, values, x.Dim)
				if err != nil {
					return make([]string, 0), err
				}
				strToUse := fmt.Sprintf(" %s ", strings.Join(xmlParamString, "\n"))
				items = append(items, strToUse)

			} else {

				paramString, err := x.Param.singleParameterXMLXX(fmt.Sprintf("%s.%s", key, x.NameToUse), x.NameToUse, values, x.Dim)
				if err != nil {
					return make([]string, 0), err
				}
				items = append(items, paramString)

			}

		}

		items = append(items, "</ds>")

	}
	return items, nil

}

// -----------------------------------------------------
//
// -----------------------------------------------------
func (p *Param) singleParameterXMLXX(key string, name string, values map[string]xmlutils.ValueDatatype, dim uint) (string, error) {
	/*
	   <data type="12p2" var="INDEC2"><![CDATA[1234567890.12]</data>
	*/

	dateType := p.GetDataType()
	specialText := p.GetSpecialText()
	dname := name

	//value := p.Value

	pramString := ""

	// if array
	if dim > 1 {

		dataList := make([]string, dim)
		for i := 0; i < int(dim); i++ {
			key := fmt.Sprintf("%s[%d]", strings.ToUpper(key), i)

			valX, found := values[key]
			valS := ""
			if found {

				valS = stringutils.AsString(valX.Value)

				if !p.HasValidValue(valS) {
					return "", fmt.Errorf("invalid value '%s' for %s (%s)", valS, key, p.DataType)
				}

			}
			dataList[i] = fmt.Sprintf("  <data type=\"%s\" var=\"%s\"  %s><![CDATA[%s]]></data>  ", dateType, dname, specialText, valS)
		}

		pramString = strings.Join(dataList, "\n")

	} else {
		valX, found := values[strings.ToUpper(key)]
		valS := ""
		if found {

			valS = stringutils.AsString(valX.Value)
			if !p.HasValidValue(valS) {
				return "", fmt.Errorf("invalid value '%s' for %s (%s)", valS, key, p.DataType)
			}
		}

		pramString = fmt.Sprintf("<data type=\"%s\" var=\"%s\" %s><![CDATA[%s]]></data>  ", dateType, dname, specialText, valS)
	}

	return pramString, nil
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
func (p *Param) GetSpecialText() string {
	speicalText, found := DataTypeSpecialText[p.DataType]
	if found {
		return speicalText
	}

	return ""
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
func (p *Param) HasValidValue(val string) bool {

	validator, found := DataTypeValidator[p.DataType]
	if found {
		return validator(val, int(p.Length), int(p.DecimalPostions))
	}

	return true
}
