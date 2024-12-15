package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/snburman/magic_game_server/config"
	"github.com/snburman/magic_game_server/db"
	"github.com/snburman/magic_game_server/handlers"
)

func main() {
	e := echo.New()
	// use cors
	e.Use(handlers.MiddlewareCORS)
	// use session middleware
	// e.Use(echo.WrapMiddleware(handlers.SessionMiddleWare))

	// serve static files
	e.Static("/", "static")

	// health check
	e.GET("/health-check", func(c echo.Context) error {
		return c.String(http.StatusOK, "service is healthy")
	})

	// auth
	authService := handlers.NewAuthService(
		config.Env().STYTCH_PROJECT_ID,
		config.Env().STYTCH_SECRET,
	)

	// password endpoints
	e.POST("/user/create", authService.HandleCreateUser)
	e.POST("/user/login", authService.HandleLoginUser)
	e.POST("/user/delete", authService.HandleDeleteUser)

	// game
	e.GET("/game", handlers.HandleGetGame)
	e.GET("/game/assets", handlers.HandleGetAssets)
	e.POST("/game/player/assets", handlers.HandleCreatePlayerAsset)
	e.GET("/game/player/assets", handlers.HandleGetPlayerAssets)

	db.NewMongoDriver()

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = ":9191"
	}
	e.Logger.Fatal(e.Start(PORT))
}
