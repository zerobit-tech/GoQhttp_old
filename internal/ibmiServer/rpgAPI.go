package ibmiServer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/zerobit-tech/GoQhttp/internal/rpg"
	"github.com/zerobit-tech/GoQhttp/internal/rpg/responseprocessor"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	"github.com/zerobit-tech/GoQhttp/logger"
	"github.com/zerobit-tech/GoQhttp/utils/httputils"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
	"github.com/zerobit-tech/GoQhttp/utils/xmlutils"
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
				ReferenceId: callID,
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

	callingXML, err := rpgEndPoint.ToXML(params)
	if err != nil {
		responseFormat = &storedProc.StoredProcResponse{
			ReferenceId: callID,
			Status:      400,
			Message:     err.Error(),
			Data:        map[string]any{},
			//LogData:     []storedProc.LogByType{{Text: fmt.Sprintf("%s", r), Type: "ERROR"}},
		}
		responseFormat.LogData = []*logger.LogEvent{logger.GetLogEvent("ERROR", callID, err.Error(), false)}
		callDuration = time.Since(t1)

		return
	}

	sqlParms := make(map[string]xmlutils.ValueDatatype)

	sqlParms["IPC"] = xmlutils.ValueDatatype{Value: "*na", DataType: "STRING"}
	sqlParms["CTL"] = xmlutils.ValueDatatype{Value: "*here *cdata", DataType: "STRING"}
	sqlParms["CI"] = xmlutils.ValueDatatype{Value: callingXML, DataType: "STRING"}

	//fmt.Println(">>>>>>>>>>>>>>>rpgPgm.ToXML(params)>>>", rpgEndPoint.ToXML(params))
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
		//fmt.Println(returnedXml.(string))
		jsonData, err := responseprocessor.ProcessSucessXML(returnedXml.(string))
		//fmt.Println(">>>>>>>>>>>>>>>response XML>>>", returnedXml.(string))
		if err == nil {

			sucessMesaage, sucessFound := jsonData["**Success"]
			if sucessFound {
				res.LogData = append(res.LogData, logger.GetLogEvent("INFO", callID, stringutils.AsString(sucessMesaage), false))
				delete(jsonData, "**Success")

				statusCodeMessage, found := jsonData["QHTTP_STATUS_MESSAGE"]
				if found {
					delete(jsonData, "QHTTP_STATUS_MESSAGE")
					res.Message = stringutils.AsString(statusCodeMessage)
				}

				statusCode, found := jsonData["QHTTP_STATUS_CODE"]
				if found {
					httpCode, message := httputils.GetValidHttpCode(statusCode)

					if httpCode > 0 {
						res.Status = httpCode

						// remove QHTTP_STATUS_CODE from out params
						delete(jsonData, "QHTTP_STATUS_CODE")

						if res.Message == "" {
							res.Message = message
						}

					}
				}

				res.Data = jsonData

			} else {
				errorData, jobErrorMDGID, err := responseprocessor.ProcessErrorXML(returnedXml.(string))
				if err == nil {
					res.Message = jobErrorMDGID
					res.Status = http.StatusBadRequest
					res.Data = map[string]any{}
					jsonString, err := json.MarshalIndent(errorData, " ", "  ")
					if err == nil {

						res.LogData = append(res.LogData, logger.GetLogEvent("ERROR", callID, string(jsonString), false))
					}

				}
			}
		}

	}

	return res, callDur, err
}
