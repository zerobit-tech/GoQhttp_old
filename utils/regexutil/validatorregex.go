package regexutil

var Regex = map[string]string{
	"EMAIL":            `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
	"URL":              `^(https?|ftp):\/\/[^\s/$.?#].[^\s]*$`,
	"NUMBER":           `^-?\d+(\.\d+)?$`,
	"POSITIVE_NUMBER":  `^\d+(\.\d+)?$`,
	"NEGATIVE_NUMBER":  `^-\d+(\.\d+)?$`,
	"POSITIVE_INTEGER": `^[1-9]\d*$`,
	"NOT_BLANK":        `^\S+$`,
	"DATE":             `^\d{4}-\d{2}-\d{2}$`,
	"TIME":             `^\d{2}:\d{2}:\d{2}$`,
	"JSON":             "__JSON__",
	"XML":              "__XML__",
}
