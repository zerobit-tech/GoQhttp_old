package xmlutils

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

type ValueDatatype struct {
	Value    any
	DataType string
}

func XmlToFlatMap(xmlString string) (map[string]any, error) {
	m := make(map[string]any)
	dtMap, _, err := XmlToFlatMapAndPlaceholder(xmlString)
	if err == nil {
		for k, v := range dtMap {
			m[k] = v.Value
		}
	}
	return m, err
}

func XmlToFlatMapAndPlaceholder(xmlString string) (map[string]ValueDatatype, string, error) {
	r := strings.NewReader(xmlString)

	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)

	elementCountMap := make(map[string]int)

	xmlMap := make(map[string]ValueDatatype)

	alreadyUsed := make([]string, 0)

	p := xml.NewDecoder(r)

	key := ""
	masterkey := ""
	for token, err := p.Token(); err == nil; token, err = p.Token() {

		switch t := token.(type) {
		case xml.CharData:
			emptyCharValue := xml.CharData([]byte(""))

			if key != "" {
				xmlMap[key] = ValueDatatype{Value: strings.TrimSpace(string([]byte(t))), DataType: "XMLSTRING"}

				x := xml.CharData([]byte(fmt.Sprintf("\"{{%s}}\"", key)))

				found := false
				for _, a := range alreadyUsed {
					if a == string(x) {
						found = true
					}
				}

				if !found {
					alreadyUsed = append(alreadyUsed, string(x))
					encoder.EncodeToken(x)
				} else {
					encoder.EncodeToken(emptyCharValue)
				}

			} else {
				encoder.EncodeToken(emptyCharValue)

			}

		case xml.EndElement:
			element_name := ""

			if t.Name.Space == "" {
				element_name = t.Name.Local

			} else {
				element_name = t.Name.Space + "." + t.Name.Local

			}

			element_name_unindexed := element_name
			count, found := elementCountMap[masterkey]
			if found && count-1 > 0 {

				element_name = fmt.Sprintf("%s[%d]", element_name, count-1)

			}

			key = strings.TrimSuffix(key, element_name)
			key = strings.TrimSuffix(key, ".")

			masterkey = strings.TrimSuffix(masterkey, element_name_unindexed)
			masterkey = strings.TrimSuffix(masterkey, ".")

			encoder.EncodeToken(t)

		case xml.StartElement:
			element_name := ""

			if t.Name.Space == "" {
				element_name = t.Name.Local

			} else {
				element_name = t.Name.Space + "." + t.Name.Local

			}
			if key == "" {
				key = element_name
			} else {
				key = key + "." + element_name
			}

			if masterkey == "" {
				masterkey = element_name
			} else {
				masterkey = masterkey + "." + element_name
			}

			e := key

			count, found := elementCountMap[e]
			if found {
				key = fmt.Sprintf("%s[%d]", key, count)
				elementCountMap[e] = count + 1
			} else {
				elementCountMap[e] = 1
			}

			placeHolderAttrs := make([]xml.Attr, 0)
			for _, a := range t.Attr {

				tempAttr := &xml.Attr{Name: a.Name}

				attribute := key + ".*ATTR"
				if a.Name.Space != "" {
					attribute = attribute + "." + a.Name.Space
				}
				attribute = attribute + "." + a.Name.Local

				xmlMap[attribute] = ValueDatatype{Value: strings.TrimSpace(a.Value), DataType: "XMLSTRING"}
				tempAttr.Value = fmt.Sprintf("\"{{%s}}\"", attribute)
				placeHolderAttrs = append(placeHolderAttrs, *tempAttr)
			}
			t.Attr = placeHolderAttrs
			encoder.EncodeToken(t)

		default:
			encoder.EncodeToken(t)

		}

	}

	// must call flush, otherwise some elements will be missing
	if err := encoder.Flush(); err != nil {
		return nil, "", err
	}

	return xmlMap, buf.String(), nil
}

//
//
//
//

func XmlToFlatMapAndPlaceholderORIGINAL(xmlString string) (map[string]ValueDatatype, string, error) {
	r := strings.NewReader(xmlString)

	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)

	elementCountMap := make(map[string]int)

	xmlMap := make(map[string]ValueDatatype)

	p := xml.NewDecoder(r)

	key := "$"
	masterkey := "$"
	for token, err := p.Token(); err == nil; token, err = p.Token() {

		switch t := token.(type) {
		case xml.CharData:

			xmlMap[key] = ValueDatatype{Value: strings.TrimSpace(string([]byte(t))), DataType: "XMLSTRING"}

			x := xml.CharData([]byte(fmt.Sprintf("{{%s}}", key)))
			encoder.EncodeToken(x)

		case xml.EndElement:
			element_name := ""

			if t.Name.Space == "" {
				element_name = "_" + t.Name.Local

			} else {
				element_name = "_" + t.Name.Space + "_" + t.Name.Local

			}

			element_name_unindexed := element_name
			count, found := elementCountMap[masterkey]
			if found && count-1 > 0 {

				element_name = fmt.Sprintf("%s[%d]", element_name, count-1)

			}

			key = strings.TrimSuffix(key, element_name)

			masterkey = strings.TrimSuffix(masterkey, element_name_unindexed)

			encoder.EncodeToken(t)

		case xml.StartElement:
			element_name := ""

			if t.Name.Space == "" {
				element_name = "_" + t.Name.Local

			} else {
				element_name = "_" + t.Name.Space + "_" + t.Name.Local

			}

			key = key + element_name
			masterkey = masterkey + element_name
			e := key

			count, found := elementCountMap[e]
			if found {
				key = fmt.Sprintf("%s[%d]", key, count)
				elementCountMap[e] = count + 1
			} else {
				elementCountMap[e] = 1
			}

			for _, a := range t.Attr {
				attribute := key + "_*ATTR_" + a.Name.Space + "_" + a.Name.Local

				xmlMap[attribute] = ValueDatatype{Value: strings.TrimSpace(a.Value), DataType: "XMLSTRING"}

			}
			encoder.EncodeToken(t)

		default:
			encoder.EncodeToken(t)

		}

	}

	// must call flush, otherwise some elements will be missing
	if err := encoder.Flush(); err != nil {
		return nil, "", err
	}

	return xmlMap, buf.String(), nil
}
