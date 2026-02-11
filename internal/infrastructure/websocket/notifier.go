package websocket

import (
	"context"
	"frauddetection/internal/domain"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSAdapter struct {
	clients map[*websocket.Conn]bool
	mu      sync.RWMutex
}

func NewWSAdapter() *WSAdapter {
	return &WSAdapter{
		clients: make(map[*websocket.Conn]bool),
	}
}

// Notify (Implementation of the Port)
func (a *WSAdapter) Notify(
	ctx context.Context,
	result domain.FraudAlert,
) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for client := range a.clients {
		err := client.WriteJSON(result)
		if err != nil {
			client.Close()
			a.removeClient(client)
		}
	}
	return nil
}

// HandleConnection (The HTTP Handler)
func (a *WSAdapter) HandleConnection(c echo.Context) {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return
	}

	// Register the new client
	a.addClient(conn)

	// Keep the connection open and listen for close events
	go func() {
		defer func() {
			a.removeClient(conn)
			conn.Close()
		}()

		for {
			//  read to detect if the connection is lost.
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

func (a *WSAdapter) addClient(conn *websocket.Conn) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.clients[conn] = true
}

func (a *WSAdapter) removeClient(conn *websocket.Conn) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.clients, conn)
}
