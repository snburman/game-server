package handlers

import (
	"github.com/labstack/echo/v4"
)

func HandleGetGame(c echo.Context) error {
	entry := []byte(

		`
		<!DOCTYPE html>
		<script src="http://localhost:9191/wasm_exec.js"></script>
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
		`)
	return c.HTMLBlob(200, entry)
}
