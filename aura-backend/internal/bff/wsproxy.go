package bff

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

func toWebSocketBase(httpBase string) string {
	httpBase = strings.TrimRight(strings.TrimSpace(httpBase), "/")
	if strings.HasPrefix(httpBase, "https://") {
		return "wss://" + strings.TrimPrefix(httpBase, "https://")
	}
	return "ws://" + strings.TrimPrefix(httpBase, "http://")
}

// ProxyInvestigationWebSocket upgrades the client connection and relays to aura-supervisor.
func (s *Server) ProxyInvestigationWebSocket(w http.ResponseWriter, r *http.Request, taskID string) {
	backendBase := toWebSocketBase(s.Cfg.SupervisorURL)
	target := backendBase + "/ws/investigations/" + url.PathEscape(taskID)
	if raw := r.URL.RawQuery; raw != "" {
		target += "?" + raw
	}

	d := websocket.Dialer{}
	backendConn, resp, err := d.Dial(target, nil)
	if err != nil {
		if resp != nil {
			w.WriteHeader(resp.StatusCode)
			return
		}
		http.Error(w, "upstream websocket unavailable", http.StatusBadGateway)
		return
	}

	clientConn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		_ = backendConn.Close()
		return
	}

	go func() {
		defer backendConn.Close()
		defer clientConn.Close()
		for {
			mt, msg, err := clientConn.ReadMessage()
			if err != nil {
				return
			}
			if err := backendConn.WriteMessage(mt, msg); err != nil {
				return
			}
		}
	}()

	for {
		mt, msg, err := backendConn.ReadMessage()
		if err != nil {
			break
		}
		if err := clientConn.WriteMessage(mt, msg); err != nil {
			break
		}
	}
}
