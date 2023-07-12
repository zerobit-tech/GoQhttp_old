package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/onlysumitg/GoQhttp/internal/dbserver"
	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/utils/concurrent"
)

func (app *application) invalidateEndPointCache() {
	app.invalidEndPointCache = true
}

func (app *application) GetEndPoint(endpoint string) (*storedProc.StoredProc, error) {
	endPoint, found := app.endPointCache[endpoint]
	if !found || app.invalidEndPointCache {
		app.endPointCache = make(map[string]*storedProc.StoredProc)
		app.endPointMutex.Lock()
		for _, sp := range app.storedProcs.List() {
			app.endPointCache[fmt.Sprintf("%s_%s", strings.ToUpper(sp.EndPointName), strings.ToUpper(sp.HttpMethod))] = sp
		}
		endPoint, found = app.endPointCache[endpoint]
		app.invalidEndPointCache = false
		app.endPointMutex.Unlock()

		if !found {

			return nil, fmt.Errorf("Not Found: %s", strings.ReplaceAll(endpoint, "_", " "))
		}

	}

	return endPoint, nil

}

var serverLastCall concurrent.MapInterface = concurrent.NewSuperEfficientSyncMap(0)

func (app *application) AddServerLastCall(serverId string) {

	serverLastCall.Store(serverId, time.Now())

}

func (app *application) ShouldPingServer(s *dbserver.Server) bool {

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
