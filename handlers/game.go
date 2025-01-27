package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/errors"
	"github.com/snburman/game_server/middleware"
	"github.com/snburman/game_server/utils"
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
	token := c.QueryParam("token")
	if id == "" || token == "" {
		return c.JSON(
			http.StatusBadRequest,
			errors.ErrMissingParams.JSON(),
		)
	}

	claims, err := utils.DecodeJWT(token)
	if err != nil || claims.UserID == "" {
		return c.JSON(
			http.StatusUnauthorized,
			errors.AuthenticationError(errors.ErrInvalidJWT).JSON(),
		)
	}

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
		`, claims.UserID))

	return c.HTMLBlob(200, entry)
}
