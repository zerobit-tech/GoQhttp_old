package mysqlserver

import (
	"strconv"
	"strings"

	"github.com/onlysumitg/GoQhttp/internal/storedProc"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func ConvertOUTVarToType(p *storedProc.StoredProcParamter, v *any) (any, error) {
	// if p.Mode != "OUT" {
	// 	return v, nil
	// }

	if v == nil {
		return nil, nil
	}

	bA, ok := (*v).([]byte)
	if !ok {
		return v, nil
	}

	strVal := string(bA)

	switch strings.ToUpper(p.Datatype) {
	case "INTEGER", "INT", "SMALLINT", "TINYINT", "MEDIUMINT", "BIGINT":
		i, err := strconv.Atoi(strVal)
		if err == nil {
			return i, nil
		} else {
			return nil, nil
		}

	case "DECIMAL", "NUMERIC":
		i, err := strconv.ParseFloat(strVal, 64)
		if err == nil {
			return i, nil
		} else {
			return nil, nil
		}

	case "FLOAT", "DOUBLE":
		i, err := strconv.ParseFloat(strVal, 32)
		if err == nil {
			return i, nil
		} else {
			return nil, nil
		}
	case "BIT":
		switch strVal {
		case "\x00":
			return false, nil
		case "\x01":
			return true, nil
		}

	}

	return strVal, nil
}
