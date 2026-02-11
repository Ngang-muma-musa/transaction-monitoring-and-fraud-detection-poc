package router

import (
	"frauddetection/internal/infrastructure/websocket"
	"net/http"

	"github.com/labstack/echo/v4"
)

type WsRouter interface {
	Register(e *echo.Echo) error
}

type wsrouter struct {
	ws *websocket.WSAdapter
}

func NewWsRouter(ws *websocket.WSAdapter) WsRouter {
	return &wsrouter{
		ws: ws,
	}
}

func (r *wsrouter) Register(e *echo.Echo) error {
	r.registerAPI(e)
	return nil
}

func (r *wsrouter) registerAPI(e *echo.Echo) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "OK",
		})
	})

	e.GET("/ws", func(c echo.Context) error {
		r.ws.HandleConnection(c)
		return nil
	})
}
