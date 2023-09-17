package ibmiServer

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/onlysumitg/GoQhttp/internal/rpg"
	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/logger"
	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
)

//-----------------------------------------------------------------

// -----------------------------------------------------------------
func (s *Server) RPGAPICall(ctx context.Context, callID string, sp *storedProc.StoredProc, rpgPgm *rpg.Program, params map[string]xmlutils.ValueDatatype, paramRegex map[string]string) (responseFormat *storedProc.StoredProcResponse, callDuration time.Duration, err error) {
	//log.Printf("%v: %v\n", "SeversCall005.001", time.Now())

	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	t1 := time.Now()
	defer func() {
		if r := recover(); r != nil {
			responseFormat = &storedProc.StoredProcResponse{
				ReferenceId: "string",
				Status:      500,
				Message:     fmt.Sprintf("%s", r),
				Data:        map[string]any{},
				//LogData:     []storedProc.LogByType{{Text: fmt.Sprintf("%s", r), Type: "ERROR"}},
			}
			responseFormat.LogData = []*logger.LogEvent{logger.GetLogEvent("ERROR", callID, fmt.Sprintf("%s", r), false)}
			callDuration = time.Since(t1)
			// apiCall.Response = responseFormat
			err = fmt.Errorf("%s", r)
		}
	}()

	sqlParms := make(map[string]xmlutils.ValueDatatype)

	sqlParms["IPC"] = xmlutils.ValueDatatype{Value: "*na", DataType: "STRING"}
	sqlParms["CTL"] = xmlutils.ValueDatatype{Value: "*here *cdata", DataType: "STRING"}
	sqlParms["CI"] = xmlutils.ValueDatatype{Value: rpgPgm.ToXML(params), DataType: "STRING"}

	fmt.Println(">>>>>>>>>>>>>>>rpgPgm.ToXML(params)>>>", rpgPgm.ToXML(params))
	givenParams := make(map[string]any)
	//.LogInfo("Building parameters for SP call")
	for k, v := range sqlParms {
		givenParams[k] = v.Value
	}
	res, callDur, err := s.call(ctx, callID, sp, givenParams, paramRegex)

	if err != nil {
		return res, callDur, err
	}

	returnedXml, found := res.Data["CO"]
	if found {
		parsedXml, err := xmlutils.XmlToFlatMap(returnedXml.(string))
		if err == nil {
			res.Data["PARSED"] = parsedXml
		}
		xx := returnedXml.(string)
		decoder := xmlutils.NewDecoder(strings.NewReader(xx))
		result, err := decoder.Decode()

		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Printf("%v\n", result)
			res.Data["PARSED2"] = result

		}
	}
	return res, callDur, err
}
