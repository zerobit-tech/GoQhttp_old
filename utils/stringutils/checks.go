package stringutils

import (
	"regexp"
	"strings"
	"time"
)

var TimeFormat string = "15:04:05"
var DateFormat string = "2006-01-02"
var ISODateFormat0 string = "20060102"

var TimestampFormat string = "2006-01-02 15:04:05.000000"
var TimestampFormat2 string = "2006-01-02 15:04:05"

func IsNumericWithOutDecimal(word string) bool {
	if strings.TrimSpace(word) == "" {
		return false
	}
	return regexp.MustCompile(`\d`).MatchString(word)
	// calling regexp.MustCompile() function to create the regular expression.
	// calling MatchString() function that returns a bool that
	// indicates whether a pattern is matched by the string.
}

func IsNumeric(word string) bool {
	if strings.TrimSpace(word) == "" {
		return false
	}
	return regexp.MustCompile(`\d*\.?\d*`).MatchString(word)
	// calling regexp.MustCompile() function to create the regular expression.
	// calling MatchString() function that returns a bool that
	// indicates whether a pattern is matched by the string.
}

func IsValidDate(stringDate string) bool {
	_, err := time.Parse(DateFormat, stringDate)
	return err == nil
}

func IsValidTime(stringDate string) bool {
	_, err := time.Parse(TimeFormat, stringDate)
	return err == nil
}

func IsValidTimeStamp(stringDate string) bool {
	_, err := time.Parse(TimestampFormat, stringDate)
	if err != nil {
		_, err := time.Parse(TimestampFormat2, stringDate)
		return err == nil
	}

	return true
}
