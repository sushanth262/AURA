package bff

import (
	"net/http"
	"net/url"
	"strings"
	"sync"

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

// ProxyInvestigationWebSocket upgrades the browser connection first, then dials aura-supervisor and relays both directions.
// Dial-before-upgrade can delay or break the client handshake on some stacks.
func (s *Server) ProxyInvestigationWebSocket(w http.ResponseWriter, r *http.Request, taskID string) {
	backendBase := toWebSocketBase(s.Cfg.SupervisorURL)
	target := backendBase + "/ws/investigations/" + url.PathEscape(taskID)
	if raw := r.URL.RawQuery; raw != "" {
		target += "?" + raw
	}

	clientConn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	d := websocket.Dialer{}
	backendConn, resp, err := d.Dial(target, nil)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		_ = clientConn.Close()
		return
	}

	var wg sync.WaitGroup
	pipe := func(dst, src *websocket.Conn) {
		defer wg.Done()
		for {
			mt, msg, readErr := src.ReadMessage()
			if readErr != nil {
				return
			}
			if writeErr := dst.WriteMessage(mt, msg); writeErr != nil {
				return
			}
		}
	}

	wg.Add(2)
	go pipe(backendConn, clientConn)
	go pipe(clientConn, backendConn)
	wg.Wait()
	_ = backendConn.Close()
	_ = clientConn.Close()
}
