package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/snburman/game-server/config"
	"github.com/snburman/game-server/errors"
	"github.com/snburman/game-server/utils"
)

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

	//TODO: refactor for connection
	// check for connection
	// _, err = conn.ConnPool.Get(claims.UserID)
	// if err != nil {
	// 	return c.JSON(
	// 		http.StatusUnauthorized,
	// 		errors.ErrConnectionNotFound.JSON(),
	// 	)
	// }

	host := config.Env().SERVER_URL
	entry := []byte(fmt.Sprintf(
		`<!DOCTYPE html>
		<link rel="stylesheet" href="/assets/styles.css">
		<script src="%s/wasm_exec.js"></script>
		<script>
			function setLoading() {
				elem = document.getElementById("loadingText");
				if (elem.innerHTML == "Loading...") {
					elem.innerHTML = "Loading.";
				} else {
					elem.innerHTML += ".";
				}
			}
			setInterval(setLoading, 500);
		</script>
		<script>function id() {return "%s"}</script>
		<script>
		if (!WebAssembly.instantiateStreaming) {
			WebAssembly.instantiateStreaming = async (resp, importObject) => {
				const source = await (await resp).arrayBuffer();
				return await WebAssembly.instantiate(source, importObjec);
				};
				}
				
				const go = new Go();
				WebAssembly.instantiateStreaming(fetch("%s/game.wasm"), go.importObject).then(result => {
					document.getElementById("loadingContainer").style.display = "none";
					go.run(result.instance);
					});
		</script>
		<div id="loadingContainer">
			<font id="loadingText">Loading...</font>
			<font id="loadingMessage">This may take a minute</font>
		</div>
		`, host, claims.UserID, host))

	return c.HTMLBlob(200, entry)
}
