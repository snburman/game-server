package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/config"
	"github.com/snburman/game_server/conn"
	"github.com/snburman/game_server/db"
	"github.com/snburman/game_server/errors"
	"github.com/snburman/game_server/utils"
)

func HandleGetGame(c echo.Context) error {
	mapID := c.Param("id")
	token := c.QueryParam("token")
	if mapID == "" || token == "" {
		return c.JSON(
			http.StatusBadRequest,
			errors.ErrMissingParams.JSON(),
		)
	}

	claims, err := utils.DecodeJWT(token)
	if err != nil || claims.UserID == "" {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ErrInvalidJWT.JSON(),
		)
	}

	// get user
	user, err := db.GetUserByID(db.MongoDB, claims.UserID)
	if err != nil {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ServerError(err.Error()).JSON(),
		)
	}
	// create Connection entry
	err = conn.NewConnection(user)
	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			errors.ServerError(err.Error()).JSON(),
		)
	}
	host := config.Env().HOST
	port := config.Env().PORT
	entry := []byte(fmt.Sprintf(
		`<!DOCTYPE html>
		<script src="%s:%s/wasm_exec.js"></script>
		<script>function id() { return %s }</script>
		<script>
		// Polyfill
		if (!WebAssembly.instantiateStreaming) {
			WebAssembly.instantiateStreaming = async (resp, importObject) => {
				const source = await (await resp).arrayBuffer();
				return await WebAssembly.instantiate(source, importObjec);
			};
		}

		const go = new Go();
		WebAssembly.instantiateStreaming(fetch("%s:%s/game.wasm"), go.importObject).then(result => {
			go.run(result.instance);
		});
		</script>`, host, port, claims.UserID, host, port))

	return c.HTMLBlob(200, entry)
}

func HandleConnectClient(c echo.Context) error {
	id := c.QueryParam("id")
	if id == "" {
		return c.JSON(
			http.StatusBadRequest,
			errors.ErrMissingParams.JSON(),
		)
	}

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

	// set connection
	conn.SetClient(id, ws)
	return c.JSON(http.StatusOK, "connected")
}
