package middleware

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/errors"
)

type WebSocketContext struct {
	echo.Context
	Ws *websocket.Conn
}

func MiddlewareWebSocket(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// create new websocket connection
		upgrader := &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return c.JSON(
				http.StatusInternalServerError,
				errors.ServerError(err.Error()).JSON(),
			)
		}

		ctx := WebSocketContext{
			Context: c,
			Ws:      ws,
		}
		return next(ctx)
	}

}
