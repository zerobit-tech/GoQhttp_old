package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zerobit-tech/GoQhttp/env"
	"github.com/zerobit-tech/GoQhttp/internal/iwebsocket"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
type GraphStats struct {
	TotalRequests int `json:"totalrequests"`

	Http100Count   int `json:"http100count"`
	Http100Percent int `json:"http100percent"`

	Http200Count   int `json:"http200count"`
	Http200Percent int `json:"http200percent"`

	Http300Count   int `json:"http300count"`
	Http300Percent int `json:"http300percent"`

	Http400Count   int `json:"http400count"`
	Http400Percent int `json:"http400percent"`

	Http500Count   int `json:"http500count"`
	Http500Percent int `json:"http500percent"`

	AvgResTime int64 `json:"avgrestime"`
	MaxResTime int64 `json:"maxrestime"`
	AvgDBTime  int64 `json:"avgdbtime"`
	MaxDBTime  int64 `json:"maxdbtime"`
}

// ------------------------------------------------------
//
// ------------------------------------------------------
type GraphStruc struct {
	Requestid string
	LogUrl    string

	Spid           string
	SpName         string
	SpUrl          string
	Httpcode       int
	HttpcodeGroup  int
	Responsetime   int64 // miliseconds
	SPResponsetime int64 // miliseconds

	Calltime string
}

// ------------------------------------------------------
//
// ------------------------------------------------------

type PlotlyMarker struct {
	Color string `json:"color"`
	Size  int    `json:"size"`
}

type GraphDatasetPlotly struct {
	X        []string     `json:"x"`
	Y        []int64      `json:"y"`
	Ploytype string       `json:"type"`
	Name     string       `json:"name"`
	Marker   PlotlyMarker `json:"marker"`
	Mode     string       `json:"mode"` // = 'lines+markers'
	Text     []string     `json:"text"`
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func getGraphDataSetPlotly(graphStruc []*GraphStruc, httpcodeGroup int) *GraphDatasetPlotly {

	gDataSet := &GraphDatasetPlotly{Ploytype: "scatter",
		Mode: "markers",
		Name: strconv.Itoa(httpcodeGroup)}

	gDataSet.X = make([]string, 0, 20)
	gDataSet.Y = make([]int64, 0, 20)

	gDataSet.Text = make([]string, 0, 20)

	for _, g := range graphStruc {
		if g == nil {
			continue
		}
		gDataSet.X = append(gDataSet.X, g.Calltime)
		gDataSet.Y = append(gDataSet.Y, g.Responsetime)

		gDataSet.Text = append(gDataSet.Text, fmt.Sprintf("%s[%s]", g.SpName, g.Requestid))
		gDataSet.Marker.Size = 12

		switch g.HttpcodeGroup {
		case 100:
			gDataSet.Marker.Color = "rgba(67, 165, 190,.7)"
		case 200:
			gDataSet.Marker.Color = "rgba(79, 176, 109,.7)"
		case 300:
			gDataSet.Marker.Color = "rgba(212, 145, 55,.7)"
		case 400:
			gDataSet.Marker.Color = "rgba(191, 44, 52,.8)"
			gDataSet.Marker.Size = 16
		case 500:
			gDataSet.Marker.Color = "rgba(240, 120, 117,.8)"
			gDataSet.Marker.Size = 16
		}
	}

	return gDataSet
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) GetGraphDataPlotyl() []*GraphDatasetPlotly {

	dataSet := make([]*GraphDatasetPlotly, 0)
	dataSet = append(dataSet, getGraphDataSetPlotly(app.GraphData100, 100))
	dataSet = append(dataSet, getGraphDataSetPlotly(app.GraphData200, 200))
	dataSet = append(dataSet, getGraphDataSetPlotly(app.GraphData300, 300))
	dataSet = append(dataSet, getGraphDataSetPlotly(app.GraphData400, 400))
	dataSet = append(dataSet, getGraphDataSetPlotly(app.GraphData500, 500))

	return dataSet

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) CaptureGraphData() {

	maxEntries, err := strconv.Atoi(env.GetEnvVariable("MAX_GRAPH_ENTRIES", "1000"))
	if err != nil || maxEntries <= 0 {
		maxEntries = 1000
	}

	//counter := 0
mainloop:
	for {

		// adding little delay to complete JS render
		//time.Sleep(500 * time.Millisecond)

		select {
		case <-app.Done:
			break mainloop
		case graphStruc, ok := <-app.GraphStream:
			if !ok {
				break mainloop
			}
			app.processGraphData(graphStruc, maxEntries)

		}

	}
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) processGraphData(graphStruc *GraphStruc, maxEntries int) {

	app.GraphStats.TotalRequests += 1

	//counter += 1

	app.graphMutex.Lock()

	httpCode := strconv.Itoa(graphStruc.Httpcode)

	if strings.HasPrefix(httpCode, "1") {
		graphStruc.HttpcodeGroup = 100
		app.GraphData100 = append([]*GraphStruc{graphStruc}, app.GraphData100...)
		if len(app.GraphData100) > maxEntries {
			app.GraphData100 = app.GraphData100[0:maxEntries]
		}
		app.GraphStats.Http100Count += 1

	}
	if strings.HasPrefix(httpCode, "2") {
		graphStruc.HttpcodeGroup = 200
		app.GraphData200 = append([]*GraphStruc{graphStruc}, app.GraphData200...)
		if len(app.GraphData200) > maxEntries {
			app.GraphData200 = app.GraphData200[0:maxEntries]
		}
		app.GraphStats.Http200Count += 1

	}

	if strings.HasPrefix(httpCode, "3") {
		graphStruc.HttpcodeGroup = 300
		app.GraphData300 = append([]*GraphStruc{graphStruc}, app.GraphData300...)
		if len(app.GraphData300) > maxEntries {
			app.GraphData300 = app.GraphData300[0:maxEntries]
		}
		app.GraphStats.Http300Count += 1

	}

	if strings.HasPrefix(httpCode, "4") {
		graphStruc.HttpcodeGroup = 400
		app.GraphData400 = append([]*GraphStruc{graphStruc}, app.GraphData400...)
		if len(app.GraphData400) > maxEntries {
			app.GraphData400 = app.GraphData400[0:maxEntries]
		}

		app.GraphStats.Http400Count += 1

	}

	if strings.HasPrefix(httpCode, "5") {
		graphStruc.HttpcodeGroup = 500
		app.GraphData500 = append([]*GraphStruc{graphStruc}, app.GraphData500...)
		if len(app.GraphData500) > maxEntries {
			app.GraphData500 = app.GraphData500[0:maxEntries]
		}
		app.GraphStats.Http500Count += 1

	}

	if app.GraphStats.Http100Count > 0 {
		app.GraphStats.Http100Percent = (app.GraphStats.Http100Count * 100 / app.GraphStats.TotalRequests)
	}

	if app.GraphStats.Http200Count > 0 {
		app.GraphStats.Http200Percent = (app.GraphStats.Http200Count * 100 / app.GraphStats.TotalRequests)
	}

	if app.GraphStats.Http300Count > 0 {
		app.GraphStats.Http300Percent = (app.GraphStats.Http300Count * 100 / app.GraphStats.TotalRequests)
	}

	if app.GraphStats.Http400Count > 0 {
		app.GraphStats.Http400Percent = (app.GraphStats.Http400Count * 100 / app.GraphStats.TotalRequests)
	}

	if app.GraphStats.Http500Count > 0 {
		app.GraphStats.Http500Percent = (app.GraphStats.Http500Count * 100 / app.GraphStats.TotalRequests)
	}

	if app.GraphStats.MaxResTime < graphStruc.Responsetime {
		app.GraphStats.MaxResTime = graphStruc.Responsetime
	}

	if app.GraphStats.MaxDBTime < graphStruc.SPResponsetime {
		app.GraphStats.MaxDBTime = graphStruc.SPResponsetime
	}

	app.GraphStats.AvgResTime = ((app.GraphStats.AvgResTime * int64(app.GraphStats.TotalRequests-1)) + graphStruc.Responsetime) / int64(app.GraphStats.TotalRequests)

	app.GraphStats.AvgDBTime = ((app.GraphStats.AvgDBTime * int64(app.GraphStats.TotalRequests-1)) + graphStruc.SPResponsetime) / int64(app.GraphStats.TotalRequests)

	response := iwebsocket.WsServerPayload{}
	response.Action = "graphdata"
	response.Message = ""
	response.Data = app.GetGraphDataPlotyl() // GetGraphData()

	go app.SendToWSChan(response)
	// go func() {
	// 	defer concurrent.Recoverer("graphdata")
	// 	iwebsocket.BroadcastToAll(response)
	// }()

	response2 := iwebsocket.WsServerPayload{}

	response2.Action = "graphtablercd"
	response2.Message = ""
	response2.Data = graphStruc
	go app.SendToWSChan(response2)

	response3 := iwebsocket.WsServerPayload{}

	response3.Action = "graphstats"
	response3.Message = ""
	response3.Data = app.GraphStats

	go app.SendToWSChan(response3)

	// go func() {
	// 	defer concurrent.Recoverer("graphtablercd")
	// 	time.Sleep(500 * time.Millisecond)

	// 	iwebsocket.BroadcastToAll(response2)
	// }()
	//time.Sleep(2 * time.Second)
	app.graphMutex.Unlock()

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) GraphHandlers(router *chi.Mux) {
	router.Route("/dashboard", func(r chi.Router) {
		//r.With(paginate).Get("/", listArticles)
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(CheckLicMiddleware)

		r.Get("/", app.GraphData)
	})
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// func GetGraphDataXX() []*GraphDataset {

// 	dataSet := make([]*GraphDataset, 0)
// 	dataSet = append(dataSet, getGraphDataSet(GraphData100, 100))
// 	dataSet = append(dataSet, getGraphDataSet(GraphData200, 200))
// 	dataSet = append(dataSet, getGraphDataSet(GraphData300, 300))
// 	dataSet = append(dataSet, getGraphDataSet(GraphData400, 400))
// 	dataSet = append(dataSet, getGraphDataSet(GraphData500, 500))

// 	return dataSet

// }

// ------------------------------------------------------
//
// ------------------------------------------------------
func minandmax(values ...int) (int, int) {
	min := values[0] //assign the first element equal to min
	max := values[0] //assign the first element equal to max
	for _, number := range values {
		if number < min {
			min = number
		}
		if number > max {
			max = number
		}
	}
	return min, max
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) GraphData(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	data.GraphData = map[int][]*GraphStruc{
		100: app.GraphData100,
		200: app.GraphData200,
		300: app.GraphData300,
		400: app.GraphData400,
		500: app.GraphData500,
	}
	//b, _ := json.Marshal(getGraphDataSet(GraphData500))
	// fmt.Println("getGraphDataSet ", getGraphDataSet(GraphData500), ":: ", string(b))
	app.render(w, r, http.StatusOK, "graph.tmpl", data)

	//app.writeJSON(w, 200, dataSet, nil)

}

/*
   label: "online tutorial subjects",
    data: [9, 8, 10, 7, 6, 12],
    backgroundColor: "red",
    borderColor: "red",
                   borderWidth: 2,

          pointRadius: 5,
*/
