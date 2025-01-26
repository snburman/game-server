package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/errors"
	"github.com/snburman/game_server/middleware"
)

func HandleAuthGame(c echo.Context) error {
	userID, err := middleware.UnmarshalClientDataContext[string](c)
	if err != nil {
		return err
	}
	fmt.Println(userID)
	return nil
}

func HandleGetGame(c echo.Context) error {
	id := c.Param("mapID")
	if id == "" {
		return c.JSON(
			http.StatusBadRequest,
			errors.ErrMissingParams.JSON(),
		)
	}

	claims, ok := c.(middleware.JWTContext)
	if !ok {
		return c.JSON(http.StatusUnauthorized, AuthResponse{
			ServerError: errors.ErrInvalidJWT,
		})
	}
	fmt.Println(claims.UserID)

	connectionID := uuid.New()

	//TODO: provide connection ID for websocket

	//TODO:

	entry := []byte(fmt.Sprintf(
		`
		<!DOCTYPE html>
		<script src="http://localhost:9191/wasm_exec.js"></script>
		<script>function id() { return %s}</script>
		<script>
		// Polyfill
		if (!WebAssembly.instantiateStreaming) {
			WebAssembly.instantiateStreaming = async (resp, importObject) => {
				const source = await (await resp).arrayBuffer();
				return await WebAssembly.instantiate(source, importObjec);
			};
		}

		const go = new Go();
		WebAssembly.instantiateStreaming(fetch("http://localhost:9191/game.wasm"), go.importObject).then(result => {
			go.run(result.instance);
		});
		</script>
		`, connectionID.String()))

	c.Response().Header().Set("CONNECTION_ID", connectionID.String())

	return c.HTMLBlob(200, entry)
}
