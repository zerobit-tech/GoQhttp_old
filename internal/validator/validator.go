package validator

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/zerobit-tech/GoQhttp/utils/xmlutils"
)

// Use the regexp.MustCompile() function to parse a regular expression pattern
// for sanity checking the format of an email address. This returns a pointer to
// a 'compiled' regexp.Regexp type, or panics in the event of an error. Parsing
// this pattern once at startup and storing the compiled *regexp.Regexp in a
// variable is more performant than re-parsing the pattern each time we need it.
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// ------------------------------------------------------
//
// ------------------------------------------------------
// Define a new Validator type which contains a map of validation errors for our
// form fields.
type Validator struct {
	FieldErrors    map[string]string `json:"field_error" db:"-" form:"-"`
	NonFieldErrors []string
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
	strconv.Quote("str")
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// Valid() returns true if the FieldErrors map doesn't contain any entries.
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// AddFieldError() adds an error message to the FieldErrors map (so long as no
// entry already exists for the given key).
func (v *Validator) AddFieldError(key, message string) {
	// Note: We need to initialize the map first, if it isn't already
	// initialized.
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}
	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}

	//v.AddNonFieldError(fmt.Sprintf("%s %s", key, message))
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// CheckField() adds an error message to the FieldErrors map only if a
// validation check is not 'ok'.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// NotBlank() returns true if a value is not an empty string.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// Must start with
func MustStartwith(value string, prefix string) bool {
	return strings.HasPrefix(strings.ToUpper(strings.Trim(value, " ")), strings.ToUpper(strings.Trim(prefix, " ")))
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// Must start with One of
func MustStartwithOneOf(value string, prefixs ...string) bool {
	for _, p := range prefixs {
		if strings.HasPrefix(strings.ToUpper(strings.Trim(value, " ")), strings.ToUpper(strings.Trim(p, " "))) {
			return true
		}
	}
	return false
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// Must start with
func MustNotStartwith(value string, prefix string) bool {
	return !strings.HasPrefix(strings.ToUpper(strings.Trim(value, " ")), strings.ToUpper(strings.Trim(prefix, " ")))
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// Cannot be
func CanNotBe(value string, prefix string) bool {
	return !strings.EqualFold(strings.ToUpper(strings.Trim(value, " ")), strings.ToUpper(strings.Trim(prefix, " ")))
}

// Matches() returns true if a value matches a provided compiled regular
// expression pattern.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// Must start with
func MustNotContainBlanks(value string) bool {
	return !strings.Contains(value, " ")
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// MinChars() returns true if a value contains at least n characters.
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// MaxChars() returns true if a value contains no more than n characters.
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// PermittedInt() returns true if a value is in a list of permitted integers.
func PermittedInt(value int, permittedValues ...int) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func MustBeFromList(value string, permittedValues ...string) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func MustBeJSON(value string) bool {

	return json.Valid([]byte(value))
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func MustBeXML(value string) bool {

	return xmlutils.IsValid(value)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func MustBeOfType(value string, typeToCheck string) bool {
	typeToCheck = strings.ToUpper(typeToCheck)

	switch typeToCheck {
	case "FLOAT64":
		_, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return false
		}
	case "INT":
		_, err := strconv.Atoi(value)
		if err != nil {
			return false
		}
	case "BOOL":
		_, err := strconv.ParseBool(value)
		if err != nil {
			return false
		}

	}

	return true
}
