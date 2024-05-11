package jsonutils

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/zerobit-tech/GoQhttp/utils/typeutils"
)

func JsonToMapPlaceholder(stringJson string) (map[string]any, error) {
	var parsedJson map[string]any
	flatmap := make(map[string]any)

	err := json.Unmarshal([]byte(stringJson), &parsedJson)
	if err == nil {
		flatmap = setMapValues(parsedJson, "")
	}

	return flatmap, err
}

// -----------------------------------------------------------
//
// -----------------------------------------------------------
func setMapValues(parsedJson map[string]any, keys string) map[string]any {

	for key, value := range parsedJson {

		keyChain := keys
		if keyChain == "" {
			keyChain = key
		} else {
			keyChain = fmt.Sprintf("%s.%s", keyChain, key) // keyChain + "__" + key
		}

		if typeutils.IsMap(value) {
			setMapValues(value.(map[string]any), keyChain)
		} else if typeutils.IsList(value) {

			listMap := make(map[string]any)
			newList, ok := value.([]any)
			if !ok {
				log.Println("err 3", value)
			}

			for i, lValue := range newList {
				xkey := fmt.Sprintf("%s[%d]", keyChain, i) // keyChain + "$$" + strconv.Itoa(i)
				listMap[xkey] = lValue

			}

			tempMap := setMapValues(listMap, "")
			tempList := make([]any, 0)
			for _, tVal := range tempMap {
				tempList = append(tempList, tVal)
			}

			parsedJson[key] = tempList

		} else {

			parsedJson[key] = fmt.Sprintf("{{%s}}", keyChain)
		}
	}

	return parsedJson

}
