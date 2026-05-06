package supervisor

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/jwtauth"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSClients multiplexes mock investigation streams by task_id.
type WSClients struct {
	mu      sync.Mutex
	clients map[string]map[*websocket.Conn]struct{}
	cfg     config.Config
}

func NewWSHub(cfg config.Config) *WSClients {
	return &WSClients{
		clients: make(map[string]map[*websocket.Conn]struct{}),
		cfg:     cfg,
	}
}

func (h *WSClients) ServeHTTP(w http.ResponseWriter, r *http.Request, taskID string) {
	token := r.URL.Query().Get("token")
	if _, err := jwtauth.VerifyBearer(r.Context(), h.cfg, "Bearer "+token); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.register(taskID, conn)
	go h.pumpReads(conn, taskID)
}

func (h *WSClients) register(taskID string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[taskID] == nil {
		h.clients[taskID] = make(map[*websocket.Conn]struct{})
	}
	h.clients[taskID][c] = struct{}{}
}

func (h *WSClients) unregister(taskID string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m, ok := h.clients[taskID]; ok {
		delete(m, c)
		if len(m) == 0 {
			delete(h.clients, taskID)
		}
	}
	_ = c.Close()
}

func (h *WSClients) pumpReads(c *websocket.Conn, taskID string) {
	defer h.unregister(taskID, c)
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			return
		}
	}
}

// Broadcast sends JSON text payload to all sockets subscribed to taskID.
func (h *WSClients) Broadcast(taskID string, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients[taskID] {
		if err := c.WriteMessage(websocket.TextMessage, b); err != nil {
			log.Printf("ws write error: %v", err)
		}
	}
}
