package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/onlysumitg/GoQhttp/internal/iwebsocket"
	"github.com/onlysumitg/GoQhttp/utils/concurrent"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) WsHandlers(router *chi.Mux) {
	router.Route("/ws", func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(CheckLicMiddleware)

		r.Get("/notification", app.WsNotification)

	})

}

var wsChan = make(chan iwebsocket.WsClientPayload, 500)

// ------------------------------------------------------
//
// ------------------------------------------------------
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) WsNotification(w http.ResponseWriter, r *http.Request) {
	// upgrade connection to websocket
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WsEndpoint", err)
	}

	log.Println("Client conneted to WsEndpoint")

	// response := &iwebsocket.WsServerPayload{Message: "Heloo", MessageType: "start"}

	// err = ws.WriteJSON(response)
	// if err != nil {
	// 	log.Println("WsEndpoint 2", err)
	// }

	conn := iwebsocket.WebSocketConnection{Conn: ws}

	// after 1st call this GoRoutine will process all websocket requests.
	go ListenForWs(&conn) //goroutine

	// ping clien
	go app.ping(&conn) //goroutine
}

// ------------------------------------------------------
//
//	get data from web socket and sent to WS channel
//
// ------------------------------------------------------@
func (app *application) ping(conn *iwebsocket.WebSocketConnection) {
	defer concurrent.Recoverer("ping")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	// ping client --> in reponse client will send pong --> check ListenToWsChannel()
	response := &iwebsocket.WsServerPayload{}
	response.Action = "ping"
	//iwebsocket.BroadcastToOne(*conn, response)

	response.Conn = conn
	//iwebsocket.BroadcastToOne(e.Conn, response)
	app.ToWSChan <- response
}

// ------------------------------------------------------
//  get data from web socket and sent to WS channel
// ------------------------------------------------------@

// ListenForWs is a goroutine that handles communication between server and client, and
// feeds data into the wsChan
func ListenForWs(conn *iwebsocket.WebSocketConnection) {

	defer concurrent.Recoverer("ListenForWs Error")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	var payload iwebsocket.WsClientPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil { // means connection closed
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ListenForWs2 error: %v", err)
			}
			break
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

// ------------------------------------------------------
//
//	get data from   WS channel and process: check main.go
//
// ------------------------------------------------------@
// ListenToWsChannel is a goroutine that waits for an entry on the wsChan, and handles it according to the
// specified action
func (app *application) ListenToWsChannel() {

	for {
		e, ok := <-wsChan

		if !ok {
			continue
		}

		//fmt.Println(">>>>>>>>>>>>>>>> WS >>>>>>>>>>>>>.", e.Action)

		switch e.Action {
		case "pong":
			// // get a list of all users and send it back via broadcast
			log.Println("Ws is ready")
			app.WSClients.Store(e.Conn, e.Username)
			// users := getUserList()
			// response.Action = "notification"
			// response.Message = "Websocket connection is sucessful."
			// response.MessageType = "success"

			// iwebsocket.BroadcastToOne(e.Conn, response)

		case "left":
			// // handle the situation where a user leaves the page
			// response.Action = "list_users"
			app.WSClients.Delete(e.Conn)
			// users := getUserList()
			// response.ConnectedUsers = users
			//iwebsocket.BroadcastToAll(response)

		case "getgraphdata":
			response := &iwebsocket.WsServerPayload{}

			response.Action = "graphdata"
			response.Message = ""
			response.Data = app.GetGraphDataPlotyl()
			response.Conn = &e.Conn
			//iwebsocket.BroadcastToOne(e.Conn, response)
			app.ToWSChan <- response

		case "getgraphstats":
			response := &iwebsocket.WsServerPayload{}

			response.Action = "graphstats"
			response.Message = ""
			response.Data = app.GraphStats
			response.Conn = &e.Conn
			//iwebsocket.BroadcastToOne(e.Conn, response)
			app.ToWSChan <- response

		case "broadcast":
			response := &iwebsocket.WsServerPayload{}

			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", e.Username, e.Message)
			//iwebsocket.BroadcastToAll(response)
			app.ToWSChan <- response
		}
	}
}

// ------------------------------------------------------@
// SendToWsChannel is a goroutine that waits for an entry on the wsChan, and handles it according to the
// specified action
// ------------------------------------------------------@

func (app *application) SendToWsChannel() {
	defer concurrent.Recoverer("SendToWsChannel")

	for {
		//time.Sleep(500 * time.Millisecond)

		dataToSend, ok := <-app.ToWSChan
		if !ok {
			continue
		}

		connectionToDelete := make([]*iwebsocket.WebSocketConnection, 0)

		if dataToSend.Conn != nil {

			err := dataToSend.Conn.WriteJSON(dataToSend)
			if err != nil {
				// the user probably left the page, or their connection dropped
				log.Println("websocket err.....................", err)
				_ = dataToSend.Conn.Close()
				connectionToDelete = append(connectionToDelete, dataToSend.Conn)
			}

		} else {
			app.WSClients.Range(func(k, v interface{}) bool {
				conn, ok := k.(iwebsocket.WebSocketConnection)
				if ok {
					err := conn.WriteJSON(dataToSend)
					if err != nil {
						// the user probably left the page, or their connection dropped
						log.Println("websocket err.....................", err)
						_ = conn.Close()
						connectionToDelete = append(connectionToDelete, &conn)
					}
				}
				return true
			})

		}

		for _, c := range connectionToDelete {
			app.WSClients.Delete(c)

		}
	}
}
