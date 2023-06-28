package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/onlysumitg/GoQhttp/env"
	"github.com/onlysumitg/GoQhttp/internal/iwebsocket"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
type GraphStruc struct {
	Requestid string
	LogUrl    string

	Spid          string
	SpName        string
	SpUrl         string
	Httpcode      int
	HttpcodeGroup int
	Responsetime  int64
	Calltime      string
}

// ------------------------------------------------------
//
// ------------------------------------------------------
type GraphXY struct {
	X string `json:"x"`
	Y int64  `json:"y"`
}

// ------------------------------------------------------
//
// ------------------------------------------------------
type GraphDataset struct {
	Label           string     `json:"label"`
	Data            []*GraphXY `json:"data"`
	BackgroundColor string     `json:"backgroundColor"`
	BorderColor     string     `json:"borderColor"`
	BorderWidth     int        `json:"borderWidth"`
	PointRadius     int        `json:"pointRadius"`
}

// ------------------------------------------------------
//
// ------------------------------------------------------

type PlotlyMarker struct {
	Color string `json:"color"`
}

type GraphDatasetPlotly struct {
	X        []string     `json:"x"`
	Y        []int64      `json:"y"`
	Ploytype string       `json:"type"`
	Name     string       `json:"name"`
	Marker   PlotlyMarker `json:"marker"`
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func getGraphDataSetPlotly(graphStruc []*GraphStruc, httpcodeGroup int) *GraphDatasetPlotly {

	gDataSet := &GraphDatasetPlotly{Ploytype: "scatter", Name: strconv.Itoa(httpcodeGroup)}
	gDataSet.X = make([]string, 0, 20)
	gDataSet.Y = make([]int64, 0, 20)

	for _, g := range graphStruc {
		if g == nil {
			continue
		}
		gDataSet.X = append(gDataSet.X, g.Calltime)
		gDataSet.Y = append(gDataSet.Y, g.Responsetime)

		switch g.HttpcodeGroup {
		case 100:
			gDataSet.Marker.Color = "#43A5BE"
		case 200:
			gDataSet.Marker.Color = "#4FB06D"
		case 300:
			gDataSet.Marker.Color = "#D49137"
		case 400:
			gDataSet.Marker.Color = "#BF2C34"
		case 500:
			gDataSet.Marker.Color = "#F07875"
		}
	}

	return gDataSet
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func GetGraphDataPlotyl() []*GraphDatasetPlotly {

	dataSet := make([]*GraphDatasetPlotly, 0)
	dataSet = append(dataSet, getGraphDataSetPlotly(GraphData100, 100))
	dataSet = append(dataSet, getGraphDataSetPlotly(GraphData200, 200))
	dataSet = append(dataSet, getGraphDataSetPlotly(GraphData300, 300))
	dataSet = append(dataSet, getGraphDataSetPlotly(GraphData400, 400))
	dataSet = append(dataSet, getGraphDataSetPlotly(GraphData500, 500))

	return dataSet

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func getGraphDataSet(graphStruc []*GraphStruc, httpcodeGroup int) *GraphDataset {

	gDataSet := &GraphDataset{BorderWidth: 2, PointRadius: 5}
	gDataSet.Label = strconv.Itoa(httpcodeGroup) ///g.Calltime.Local().Format(TimestampFormat)

	for _, g := range graphStruc {
		if g == nil {
			continue
		}
		gDataSet.Data = append(gDataSet.Data, &GraphXY{X: g.Calltime, Y: g.Responsetime})

		switch g.HttpcodeGroup {
		case 100:
			gDataSet.BorderColor = "blue"
		case 200:
			gDataSet.BorderColor = "green"
		case 300:
			gDataSet.BorderColor = "yellow"
		case 400:
			gDataSet.BorderColor = "red"
		case 500:
			gDataSet.BorderColor = "purple"
		}
		gDataSet.BackgroundColor = gDataSet.BorderColor
	}

	return gDataSet
}

// ------------------------------------------------------
//
// ------------------------------------------------------
var GraphData100 []*GraphStruc = make([]*GraphStruc, 0, 200)
var GraphData200 []*GraphStruc = make([]*GraphStruc, 0, 500)
var GraphData300 []*GraphStruc = make([]*GraphStruc, 0, 200)
var GraphData400 []*GraphStruc = make([]*GraphStruc, 0, 200)
var GraphData500 []*GraphStruc = make([]*GraphStruc, 0, 500)

// ------------------------------------------------------
//
// ------------------------------------------------------
var GraphChan chan *GraphStruc = make(chan *GraphStruc, 5000)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) CaptureGraphData() {

	maxEntries, err := strconv.Atoi(env.GetEnvVariable("MAX_GRAPH_ENTRIES", "1000"))
	if err != nil || maxEntries <= 0 {
		maxEntries = 1000
	}

	for {
		graphStruc := <-GraphChan

		httpCode := strconv.Itoa(graphStruc.Httpcode)

		if strings.HasPrefix(httpCode, "1") {
			graphStruc.HttpcodeGroup = 100
			GraphData100 = append([]*GraphStruc{graphStruc}, GraphData100...)
			if len(GraphData100) > maxEntries {
				GraphData100 = GraphData100[0:maxEntries]
			}

		}
		if strings.HasPrefix(httpCode, "2") {
			graphStruc.HttpcodeGroup = 200
			GraphData200 = append([]*GraphStruc{graphStruc}, GraphData200...)
			if len(GraphData200) > maxEntries {
				GraphData200 = GraphData200[0:maxEntries]
			}
		}

		if strings.HasPrefix(httpCode, "3") {
			graphStruc.HttpcodeGroup = 300
			GraphData300 = append([]*GraphStruc{graphStruc}, GraphData300...)
			if len(GraphData300) > maxEntries {
				GraphData300 = GraphData300[0:maxEntries]
			}
		}

		if strings.HasPrefix(httpCode, "4") {
			graphStruc.HttpcodeGroup = 400
			GraphData400 = append([]*GraphStruc{graphStruc}, GraphData400...)
			if len(GraphData400) > maxEntries {
				GraphData400 = GraphData400[0:maxEntries]
			}
		}

		if strings.HasPrefix(httpCode, "5") {
			graphStruc.HttpcodeGroup = 500
			GraphData500 = append([]*GraphStruc{graphStruc}, GraphData500...)
			if len(GraphData500) > maxEntries {
				GraphData500 = GraphData500[0:maxEntries]
			}
		}

		var response iwebsocket.WsServerPayload
		response.Action = "graphdata"
		response.Message = ""
		response.Data = GetGraphDataPlotyl() // GetGraphData()
		go func() {
			defer func() {
				if r := recover(); r != nil {
				}
			}()
			iwebsocket.BroadcastToAll(response)
		}()

		var response2 iwebsocket.WsServerPayload

		response2.Action = "graphtablercd"
		response2.Message = ""
		response2.Data = graphStruc
		go func() {
			defer func() {
				if r := recover(); r != nil {
				}
			}()
			iwebsocket.BroadcastToAll(response2)
		}()
		//time.Sleep(2 * time.Second)

	}
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) GraphHandlers(router *chi.Mux) {
	router.Route("/graph", func(r chi.Router) {
		//r.With(paginate).Get("/", listArticles)
		r.Use(app.RequireAuthentication)
		r.Get("/", app.GraphData)
	})
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func GetGraphDataXX() []*GraphDataset {

	dataSet := make([]*GraphDataset, 0)
	dataSet = append(dataSet, getGraphDataSet(GraphData100, 100))
	dataSet = append(dataSet, getGraphDataSet(GraphData200, 200))
	dataSet = append(dataSet, getGraphDataSet(GraphData300, 300))
	dataSet = append(dataSet, getGraphDataSet(GraphData400, 400))
	dataSet = append(dataSet, getGraphDataSet(GraphData500, 500))

	return dataSet

}

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
		100: GraphData100,
		200: GraphData200,
		300: GraphData300,
		400: GraphData400,
		500: GraphData500,
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
