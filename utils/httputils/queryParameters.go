package httputils

import (
	"encoding/json"
	"net/url"
	"strings"
)

func QueryParamPath(urlString string, removePrefix string) (string, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return urlString, err
	}

	path := u.Path

	if removePrefix != "" {
		path = strings.TrimPrefix(path, removePrefix)

	}

	path = strings.Trim(path, "/")
	return path, err
}

func QueryParamToMap(urlString string) (map[string]any, error) {
	asRequestJson := make(map[string]any)
	u, err := url.Parse(urlString)
	if err != nil {
		return asRequestJson, err
	}

	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return asRequestJson, err
	}

	for k, v := range q {
		k := strings.ToUpper(k)
		switch len(v) {
		case 0:
			asRequestJson[k] = ""
		case 1:
			asRequestJson[k] = v[0]
		default:
			asRequestJson[k] = v
		}

	}

	return asRequestJson, nil

}
func QueryParamToJson(urlString string) (string, error) {

	asRequestJson, err := QueryParamToMap(urlString)
	if err != nil {
		return "", err
	}
	jsonStr, err := json.Marshal(asRequestJson)

	return string(jsonStr), err
}
