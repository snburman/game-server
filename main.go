package main

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/config"
	"github.com/snburman/game_server/db"
	"github.com/snburman/game_server/handlers"
)

func main() {
	e := echo.New()
	// use cors
	e.Use(handlers.MiddlewareCORS)

	// use session
	store := sessions.NewCookieStore([]byte(config.Env().SECRET))
	e.Use(session.Middleware(store))

	// auth
	authService := handlers.NewAuthService(store)

	// serve static files
	e.Static("/", "static")

	// health check
	e.GET("/health-check", func(c echo.Context) error {
		return c.String(http.StatusOK, "service is healthy")
	})

	// token refresh
	e.POST("/token/refresh", authService.HandleRefreshToken)

	// user endpoints
	e.GET("/user", authService.HandleGetUser)
	e.POST("/user/create", authService.HandleCreateUser)
	e.POST("/user/login", authService.HandleLoginUser)
	e.DELETE("/user/delete", authService.HandleDeleteUser)

	// game
	e.GET("/game", handlers.HandleGetGame)

	// assets
	e.GET("/assets", handlers.HandleGetAssets)
	e.POST("/assets/player", handlers.HandleCreatePlayerAsset)
	e.GET("/assets/player", handlers.HandleGetPlayerAssets)

	db.NewMongoDriver()

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = ":9191"
	}
	e.Logger.Fatal(e.Start(PORT))
}
