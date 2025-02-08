package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/snburman/game-server/config"
	"github.com/snburman/game-server/conn"
	"github.com/snburman/game-server/db"
	"github.com/snburman/game-server/errors"
	"github.com/snburman/game-server/middleware"
	"github.com/snburman/game-server/utils"
)

func HandleClientConnect(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
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

	ctx, ok := c.(middleware.WebSocketContext)
	if !ok {
		log.Println("missing websocket context")
		return c.JSON(
			http.StatusInternalServerError,
			errors.ErrServerError.JSON(),
		)
	}

	// set connection
	err = conn.SetClient(claims.UserID, ctx.Ws)
	if err != nil {
		ctx.Ws.Close()
		return c.NoContent(http.StatusInternalServerError)
	}

	conn, err := conn.ConnPool.Get(claims.UserID)
	if err != nil {
		log.Println(err)
		ctx.Ws.Close()
		return c.JSON(
			http.StatusInternalServerError,
			errors.ErrConnectionNotFound.JSON(),
		)
	}
	conn.Client.Write("hello")
	for {
		// read message
		msg, err := conn.Client.Read()
		if err != nil {
			log.Println(err)
			break
		}
		conn.Client.Write(msg)
	}
	return nil
}

func HandleGetGame(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
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

	// check for connection
	_, err = conn.ConnPool.Get(claims.UserID)
	if err != nil {
		return c.JSON(
			http.StatusUnauthorized,
			errors.ErrConnectionNotFound.JSON(),
		)
	}

	host := config.Env().HOST
	port := config.Env().PORT
	entry := []byte(fmt.Sprintf(
		`<!DOCTYPE html>
		<script src="%s:%s/wasm_exec.js"></script>
		<script>function id() {return "%s"}</script>
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
