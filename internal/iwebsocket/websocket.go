package iwebsocket

import (
	"github.com/gorilla/websocket"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
// WebSocketConnection is a wrapper for our websocket connection, in case
// we ever need to put more data into the struct
type WebSocketConnection struct {
	*websocket.Conn
}

// WsPayload defines the websocket request from the client
// ------------------------------------------------------
//
//	payload from client to server
//
// ------------------------------------------------------
type WsClientPayload struct {
	Action   string              `json:"action"`
	Username string              `json:"username"`
	Message  string              `json:"message"`
	Conn     WebSocketConnection `json:"-"`
}

// ------------------------------------------------------
// list of clients
// ------------------------------------------------------
//var Clients = make(map[WebSocketConnection]string)

//var Clients concurrent.MapInterface = concurrent.NewSuperEfficientSyncMap(0)

// ------------------------------------------------------
//
//	payload from server to client
//
// ------------------------------------------------------
type WsServerPayload struct {
	Action      string               `json:"action"`
	Message     string               `json:"message"`
	MessageType string               `json:"messagetype"`
	Data        any                  `json:"data"`
	Conn        *WebSocketConnection `json:"-"`
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// broadcastToAll sends ws response to all connected clients
// func BroadcastNotification(message, messageType string) {
// 	serverPayload := &WsServerPayload{
// 		Action:      "notification",
// 		Message:     message,
// 		MessageType: messageType,
// 	}
// 	BroadcastToAll(serverPayload)
// }

// ------------------------------------------------------
//
// ------------------------------------------------------
// broadcastToAll sends ws response to all connected clients
// func BroadcastToAll(response *WsServerPayload) {

// 	connectionToDelete := make([]WebSocketConnection, 0)

// 	Clients.Range(func(k, v interface{}) bool {
// 		conn, ok := k.(WebSocketConnection)
// 		if ok {
// 			err := conn.WriteJSON(response)
// 			if err != nil {
// 				// the user probably left the page, or their connection dropped
// 				log.Println("websocket err.....................", err)
// 				 _ = conn.Close()
// 				connectionToDelete = append(connectionToDelete, conn)
// 			}
// 		}
// 		return true
// 	})

// 	for _, c := range connectionToDelete {
// 		Clients.Delete(c)

// 	}

// }

// ------------------------------------------------------
//
// ------------------------------------------------------
// broadcastToAll sends ws response to all connected clients
// func BroadcastToOne(conn WebSocketConnection, response *WsServerPayload) {

// 	err := conn.WriteJSON(response)
// 	if err != nil {
// 		// the user probably left the page, or their connection dropped
// 		log.Println("websocket err")
// 		_ = conn.Close()
// 		Clients.Delete(conn)
// 	}

// }
