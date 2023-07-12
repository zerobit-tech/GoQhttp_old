package logger

import "regexp"

var NoLogRegex map[string]*regexp.Regexp = map[string]*regexp.Regexp{
	"CREDIT CARD": regexp.MustCompile(`(?mi)(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\\d{3})\\d{11})`),
	"AMEX CARD":   regexp.MustCompile(`(?mi)3[47][0-9]{13}`),
	//"DISCOVER CARD": regexp.MustCompile(`(?mi)65[4-9][0-9]{13}|64[4-9][0-9]{13}|6011[0-9]{12}|(622(?:12[6-9]|1[3-9][0-9]|[2-8][0-9][0-9]|9[01][0-9]|92[0-5])[0-9]{10})`),
	//"MASTERCARD":    regexp.MustCompile(`(?mi)(5[1-5][0-9]{14}|2(22[1-9][0-9]{12}|2[3-9][0-9]{13}|[3-6][0-9]{14}|7[0-1][0-9]{13}|720[0-9]{12}))`),
	//"VISA CARD":     regexp.MustCompile(`(?mi)4[0-9]{12}(?:[0-9]{3})?`),
	"SSN": regexp.MustCompile(`(?mi)\d{3}[- ]?\d{2}[- ]?\d{4}`),
}

func RemoveNonLogData(data string) string {
	d := data
	for k, r := range NoLogRegex {
		d = r.ReplaceAllString(d, k)
	}
	return d
}
