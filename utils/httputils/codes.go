package httputils

import "net/http"

func IsValidHttpCode(code int) (valid bool, message string) {
	message = http.StatusText(code)

	valid = (message != "")

	return
}
