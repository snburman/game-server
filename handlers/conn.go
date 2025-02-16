package handlers

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/snburman/game-server/conn"
)

func HandleGameWebsocket(c echo.Context) error {
	userID := c.Param("userID")
	conn, err := conn.NewConn(c.Response(), c.Request(), userID)
	if err != nil {
		return c.JSON(500, err)
	}
	log.Println("new connection created: ", conn.ID)
	conn.Listen()

	return nil
}
