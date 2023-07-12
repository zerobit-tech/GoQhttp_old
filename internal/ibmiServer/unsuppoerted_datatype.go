package ibmiServer

var unsupportedDataType []string = []string{
	"BINARY",
	"BINARY VARYING",
	"BINARY LARGE OBJECT",

	"GRAPHIC",
	"GRAPHIC VARYING",
	"DOUBLE-BYTE CHARACTER LARGE OBJECT",
}

var unsupportedOUTDataType []string = []string{
	"ROWID",
}

func isUnsupportedDataType(name string, usage string) bool {
	for _, t := range unsupportedDataType {
		if t == name {
			return true
		}
	}

	if usage == "OUT" {
		for _, t := range unsupportedOUTDataType {
			if t == name {
				return true
			}
		}
	}
	return false
}
