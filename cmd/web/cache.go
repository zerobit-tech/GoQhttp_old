package main

import (
	"errors"
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/zerobit-tech/GoQhttp/env"
	"github.com/zerobit-tech/GoQhttp/internal/ibmiServer"
	"github.com/zerobit-tech/GoQhttp/internal/models"
	"github.com/zerobit-tech/GoQhttp/internal/rpg"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
	"github.com/zerobit-tech/GoQhttp/utils/regexutil"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) invalidateEndPointCache() {
	app.invalidEndPointCache = true
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) GetEndPoint(namespace, endpointName, httpmethod string) (*storedProc.StoredProc, error) {
	endPointKey := fmt.Sprintf("%s_%s_%s", strings.ToUpper(namespace), strings.ToUpper(endpointName), strings.ToUpper(httpmethod))

	endPoint, found := app.endPointCache[endPointKey]
	if !found || app.invalidEndPointCache {
		app.endPointCache = make(map[string]*storedProc.StoredProc)
		app.endPointMutex.Lock()
		for _, sp := range app.storedProcs.List(true) {
			app.endPointCache[fmt.Sprintf("%s_%s_%s", strings.ToUpper(sp.Namespace), strings.ToUpper(sp.EndPointName), strings.ToUpper(sp.HttpMethod))] = sp
		}
		endPoint, found = app.endPointCache[endPointKey]
		app.invalidEndPointCache = false
		app.endPointMutex.Unlock()

		if !found {

			return nil, fmt.Errorf("Not Found: %s", strings.ReplaceAll(endPointKey, "_", " "))
		}

	}

	return endPoint, nil

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) GetRPGEndPoint(namespace, endpointName, httpmethod string) (*rpg.RpgEndPoint, error) {

	rpgEndPointId := fmt.Sprintf("%s_%s_%s", strings.ToLower(namespace), strings.ToLower(endpointName), strings.ToLower(httpmethod))

	rpgEndPoint, err := app.RpgEndpointModel.Get(rpgEndPointId)

	return rpgEndPoint, err

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) GetRPGDriver(server *ibmiServer.Server) (*storedProc.StoredProc, error) {

	sp, err := app.GetEndPoint(env.RpgDriverNameSpace(server.Name), env.RpgDefaultDriverprogram(server.Name), "post") //iPLUGR512K

	if err != nil {
		return nil, errors.New("RPG Driver not found!")
	}

	return sp, err

}

// ------------------------------------------------------
//
// ------------------------------------------------------
var serverLastCall concurrent.MapInterface = concurrent.NewSuperEfficientSyncMap(0)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) AddServerLastCall(serverId string) {

	serverLastCall.Store(serverId, time.Now())

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) ShouldPingServer(s *ibmiServer.Server) bool {

	lastCall, found := serverLastCall.Load(s.ID)

	if !found {
		return true
	}

	lastCallTime, ok := lastCall.(time.Time)

	if !ok {
		return true
	}

	idleDuration := time.Duration(s.ConnectionIdleAge) * time.Second
	//fmt.Println("time.Since(lastCallTime)", time.Since(lastCallTime), "::", idleDuration)
	return (time.Since(lastCallTime) >= idleDuration)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) GetParamValidatorRegex() map[string]string {
	return app.paramRegexModel.Map()
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) LoadDefaultParamValidatorRegex() {

	defer concurrent.Recoverer("LoadDefaultParamValidatorRegex")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	for k, v := range regexutil.Regex {
		key := strings.ToUpper(k)
		_, err := regexp.Compile(v)
		if err == nil {

			_, err2 := app.paramRegexModel.Get(key)
			if err2 != nil { // if not found ==> add it
				rp := &models.ParamRegex{
					Name:  key,
					Regex: v,
				}
				app.paramRegexModel.Save(rp)

			}

		}

	}
}
