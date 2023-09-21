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
	"github.com/onlysumitg/GoQhttp/utils/stringutils"
	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
)

//-----------------------------------------------------------------

// -----------------------------------------------------------------
func (s *Server) RPGAPICall(ctx context.Context, callID string, sp *storedProc.StoredProc, rpgEndPoint *rpg.RpgEndPoint, params map[string]xmlutils.ValueDatatype, paramRegex map[string]string) (responseFormat *storedProc.StoredProcResponse, callDuration time.Duration, err error) {
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
	sqlParms["CI"] = xmlutils.ValueDatatype{Value: rpgEndPoint.RpgPgm.ToXML(params), DataType: "STRING"}

	fmt.Println(">>>>>>>>>>>>>>>rpgPgm.ToXML(params)>>>", rpgEndPoint.RpgPgm.ToXML(params))
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

		xx := returnedXml.(string)
		decoder := xmlutils.NewDecoder(strings.NewReader(xx))
		result, err := decoder.Decode()

		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			//	fmt.Printf("%v\n", result)
			finalValues := make(map[string]any)

			xmlservice, found := result["xmlservice"]
			if found {

				xmlserviceMap, ok := xmlservice.(map[string]any)

				if ok {
					pgm, found := xmlserviceMap["pgm"]
					if found {

						pgmMap, ok := pgm.(map[string]any)
						if ok {
							parms, found := pgmMap["parm"]
							if found {
								parmsList, ok := parms.([]map[string]any)
								if ok {
									for _, parms := range parmsList {

										varName, foundName := parms["@var"]

										data, found := parms["data"]
										if found {
											dataMap, ok := data.(map[string]any)
											if ok {
												val, found := dataMap["#text"]
												if found && foundName {
													finalValues[stringutils.AsString(varName)] = val
												}
											}
										}

									}

								}
							}
						}
					}

				}

			}
			res.Data["PARSED2"] = finalValues

		}
	}
	return res, callDur, err
}
