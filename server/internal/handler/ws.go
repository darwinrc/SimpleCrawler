package handler

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"server/internal/model"
	"server/internal/service"
)

type WsHandler interface {
	Attach(r *mux.Router)
	HandleWebSocketConnection(w http.ResponseWriter, r *http.Request)
	//ProcessCrawledUrls(ctx context.Context)
}

type wsHandler struct {
	Service service.CrawlerService
	Context context.Context
}

// NewWsHandler builds a handler and injects its dependencies
func NewWsHandler(s service.CrawlerService) WsHandler {
	return &wsHandler{
		Service: s,
	}
}

// Attach attaches the websocket endpoint to the router
func (h *wsHandler) Attach(r *mux.Router) {
	r.HandleFunc("/ws", h.HandleWebSocketConnection)
}

// HandleWebSocketConnection establishes a web socket connection and reads messages coming through it
func (h *wsHandler) HandleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading connection to support websockets: %s", err)
		return
	}

	h.Context = r.Context()

	h.readMessages(conn)
}

// clients holds the list of websocket connections mapped with the corresponding request id
var clients = make(map[*websocket.Conn]string)

// readMessages watches for messages coming through the websocket connection
func (h *wsHandler) readMessages(conn *websocket.Conn) {
	defer conn.Close()

	log.Printf("new connection: %s", conn.RemoteAddr().String())

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error getting reader: %s", err)
		}
		reqId := uuid.New().String()

		req := &model.Request{
			ReqId: reqId,
		}

		if err := json.Unmarshal(msg, req); err != nil {
			log.Printf("error unmarshaling request: %s", err)
			return
		}

		clients[conn] = req.ReqId

		// Separate goroutine for writing the crawled urls to the websocket connection
		broadcast := make(chan []byte)
		go h.ProcessCrawledUrls(broadcast)

		h.Service.Crawl(h.Context, reqId, req.Url, broadcast)
	}
}

// ProcessCrawledUrls watches for messages in the broadcast channel and send them to the corresponding clients
func (h *wsHandler) ProcessCrawledUrls(broadcast chan []byte) {

	for {
		msg := <-broadcast

		for client, reqId := range clients {
			res := &model.Response{}
			if err := json.Unmarshal(msg, res); err != nil {
				log.Printf("error unmarshaling response: %s", err)
				return
			}

			if reqId != res.ReqId {
				continue
			}

			if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
				delete(clients, client)
				client.Close()
			}
		}
	}
}
