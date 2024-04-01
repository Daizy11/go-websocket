package main

import (
    "fmt"
    "github.com/gorilla/websocket"
    gubrak "github.com/novalagung/gubrak/v2"
    "os"
    "log"
    "net/http"
    "strings"
)

type M map[string]interface{}

type SocketPayload struct { //menampung payload
	Message string
}

type SocketResponse struct { //digunakan oleh back end (socket server) sewaktu membroadcast message ke semua client yang terhubung
	From    string
	Type    string
	Message string
}

type WebSocketConnection struct {
	*websocket.Conn
	Username string
}

const MESSAGE_NEW_USER = "New User"
const MESSAGE_CHAT = "Chat"
const MESSAGE_LEAVE = "Leave"

var connections = make([]*WebSocketConnection, 0)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, err := os.ReadFile("index.html")
		if err != nil {
			http.Error(w, "Could not open requested file", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%s", content)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		currentGorillaConn, err := upgrader.Upgrade(w, r, w.Header())
		if err != nil {
			http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		}

		username := r.URL.Query().Get("username")
		currentConn := WebSocketConnection{Conn: currentGorillaConn, Username: username}
		connections = append(connections, &currentConn)

		go handleIO(&currentConn, connections)
	})
	fmt.Println("Server starting at :8080")
	http.ListenAndServe(":8080", nil)
}

func handleIO(currentConn *WebSocketConnection, connections []*WebSocketConnection) {
	//Proses broadcast message ke semua client yg terhubung dilakukan dalam fungsi ini.
    
	defer func() { //defer = eksekusi terakhir
        if r := recover(); r != nil {
            log.Println("ERROR", fmt.Sprintf("%v", r))
        }
    }()

    broadcastMessage(currentConn, MESSAGE_NEW_USER, "")

    for {
        payload := SocketPayload{}
        err := currentConn.ReadJSON(&payload)
        if err != nil {
            if strings.Contains(err.Error(), "websocket: close") {
                broadcastMessage(currentConn, MESSAGE_LEAVE, "")
                ejectConnection(currentConn)
                return
            }

            log.Println("ERROR", err.Error())
            continue
        }

        broadcastMessage(currentConn, MESSAGE_CHAT, payload.Message)
    }
}

 func broadcastMessage(currentConn *WebSocketConnection, kind, message string) {
     for _, eachConn := range connections {
		 fmt.Println(eachConn)
         if eachConn == currentConn {
             continue
         }

         eachConn.WriteJSON(SocketResponse{
             From:    currentConn.Username,
             Type:    kind,
             Message: message,
         })
     }
 }

 func ejectConnection(currentConn *WebSocketConnection) {
     filtered := gubrak.From(connections).Reject(func(each *WebSocketConnection) bool {
         return each == currentConn
     }).Result()
     connections = filtered.([]*WebSocketConnection)
 }
