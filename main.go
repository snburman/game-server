package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/snburman/magic_game_server/db"
	"github.com/snburman/magic_game_server/handlers"
)

func main() {
	e := echo.New()
	// use cors
	e.Use(handlers.MiddlewareCORS)
	// use session middleware
	e.Use(echo.WrapMiddleware(handlers.SessionMiddleWare))

	// serve static files
	e.Static("/", "static")

	// health check
	e.GET("/health-check", func(c echo.Context) error {
		return c.String(http.StatusOK, "service is healthy")
	})

	// auth
	e.GET("/login", handlers.HandleLogin)
	e.POST("/auth", handlers.HandleAuthenticate)
	e.POST("/register", handlers.HandleRegister)
	e.GET("/logout", handlers.HandleLogout)

	// session
	e.GET("/session/create", handlers.HandleCreateSession)
	e.GET("/session/get", handlers.HandleGetSession)
	e.GET("/session/find", handlers.HandleFindSession)

	// game
	e.GET("/game", handlers.HandleGetGame)
	e.GET("/game/assets", handlers.HandleGetAssets)
	e.POST("/game/player/assets", handlers.HandleCreatePlayerAsset)
	e.GET("/game/player/assets", handlers.HandleGetPlayerAssets)

	db.NewMongoDriver()
	e.Logger.Fatal(e.Start(":9191"))
}
