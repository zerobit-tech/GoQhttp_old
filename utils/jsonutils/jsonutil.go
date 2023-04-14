package jsonutils

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"

	"github.com/onlysumitg/GoQhttp/utils/typeutils"
	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
)

/*
	{
		"a": "aval",
		"b": {
			"x1": "x1val",
			"X2": {
				"x3": "x3val"
			},
			"x4": ["x41", "x42", "x43"]
		},
		"c": ["c1", "c2", "c3"],
		"d": true,
	        "dd":1,
	        "dd2":3.20,

"e": null
}

# TO

map[a:{aval string}

	b.X2.x3:{x3val string}
	b.x1:{x1val string}
	b.x4[0]:{x41 string}
	b.x4[1]:{x42 string}
	b.x4[2]:{x43 string}
	c[0]:{c1 string}
	c[1]:{c2 string}
	c[2]:{c3 string}
	d:{true bool}
	dd:{1 float64}
	dd2:{3.2 float64}
	e:{<nil> invalid}]
*/
type JsonValues_NOTINTUSE struct {
	Value    any
	DataType string
}

// -----------------------------------------------------------
//
// -----------------------------------------------------------
func JsonToFlatMap(stringJson string) (map[string]xmlutils.ValueDatatype, error) {
	var parsedJson map[string]any
	flatmap := make(map[string]xmlutils.ValueDatatype)

	err := json.Unmarshal([]byte(stringJson), &parsedJson)
	if err == nil {
		flatmap = processValue(parsedJson, "")
	}

	return flatmap, err
}

// -----------------------------------------------------------
//
// -----------------------------------------------------------
func JsonToFlatMapFromMap(parsedJson map[string]any) map[string]xmlutils.ValueDatatype {

	flatmap := processValue(parsedJson, "")

	return flatmap
}

// -----------------------------------------------------------
//
// -----------------------------------------------------------
func processValue(value any, keyChain string) map[string]xmlutils.ValueDatatype {
	flatmap := make(map[string]xmlutils.ValueDatatype)
	if typeutils.IsMap(value) {
		newValueMap, ok := value.(map[string]any)
		if !ok {
			log.Println("ERRORRRRRRRRRRRRRR 1", value)
		} else {
			iMap := buildFlatMap(newValueMap, keyChain)

			for ikey, ivalue := range iMap {
				flatmap[ikey] = ivalue
			}
		}
	} else if typeutils.IsList(value) {
		newList, ok := value.([]any)
		if !ok {
			log.Println("ERRORRRRRRRRRRRRRR 2", value)
		}

		iMap := buildFlatList(newList, keyChain)

		for ikey, ivalue := range iMap {
			flatmap[ikey] = ivalue
		}

	} else {

		flatmap[keyChain] = xmlutils.ValueDatatype{Value: value, DataType: fmt.Sprint(reflect.ValueOf(value).Kind())}

	}
	return flatmap

}

//-----------------------------------------------------------
//
//-----------------------------------------------------------

func buildFlatList(jsonArray []any, keys string) map[string]xmlutils.ValueDatatype {
	flatmap := make(map[string]xmlutils.ValueDatatype)

	for i, val := range jsonArray {

		keyChain := keys
		if keyChain == "" {
			keyChain = strconv.Itoa(i)
		} else {
			keyChain = fmt.Sprintf("%s[%d]", keyChain, i) // keyChain + "$$" + strconv.Itoa(i)
		}

		iMap := processValue(val, keyChain)
		for ikey, ivalue := range iMap {
			flatmap[ikey] = ivalue
		}
	}

	return flatmap
}

// -----------------------------------------------------------
//
// -----------------------------------------------------------
func buildFlatMap(parsedJson map[string]any, keys string) map[string]xmlutils.ValueDatatype {
	flatmap := make(map[string]xmlutils.ValueDatatype)

	for key, value := range parsedJson {

		keyChain := keys
		if keyChain == "" {
			keyChain = key
		} else {
			keyChain = fmt.Sprintf("%s.%s", keyChain, key) // keyChain + "__" + key
		}
		iMap := processValue(value, keyChain)
		for ikey, ivalue := range iMap {
			flatmap[ikey] = ivalue
		}
	}

	return flatmap

}
