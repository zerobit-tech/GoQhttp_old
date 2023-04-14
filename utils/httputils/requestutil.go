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
