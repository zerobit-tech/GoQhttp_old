package main

import (
	"fmt"
	"strings"

	"github.com/onlysumitg/GoQhttp/internal/models"
)

func (app *application) invalidateEndPointCache() {
	app.invalidEndPointCache = true
}

func (app *application) GetEndPoint(endpoint string) (*models.StoredProc, error) {
	endPoint, found := app.endPointCache[endpoint]
	if !found || app.invalidEndPointCache {
		app.endPointCache = make(map[string]*models.StoredProc)
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
