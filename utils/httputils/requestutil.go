package httputils

import (
	"net/http"
	"strings"
)

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}

func GetHeadersAsMap(r *http.Request) map[string]string {

	returnMap := make(map[string]string)
	for name, values := range r.Header {
		// Loop over all values for the name.
		returnMap[name] = strings.Join(values, ",")

	}

	return returnMap
}

func GetHeadersAsMap2(h http.Header) map[string]string {

	returnMap := make(map[string]string)
	for name, values := range h {
		// Loop over all values for the name.
		returnMap[name] = strings.Join(values, ",")

	}

	return returnMap
}

func HasFormData(r *http.Request) bool {

	contentType := r.Header.Get("Content-Type")

	return strings.EqualFold(contentType, "application/x-www-form-urlencoded")

}

func FormToJson(r *http.Request) (map[string]any, error) {
	asJson := make(map[string]any)
	err := r.ParseForm()
	if err != nil {
		return asJson, err
	}

	for k, v := range r.Form {
		kU := strings.ToUpper(k)
		length := len(v)

		switch length {
		case 0:
			asJson[kU] = ""
		case 1:
			asJson[kU] = v[0]
		default:
			asJson[kU] = v
		}

	}

	return asJson, nil
}
