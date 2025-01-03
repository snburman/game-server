package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/db"
	"github.com/snburman/game_server/handlers"
	"github.com/snburman/game_server/middleware"
)

func main() {
	e := echo.New()
	// use cors
	e.Use(middleware.MiddlewareCORS)

	// auth
	authService := handlers.NewAuthService()

	// serve static files
	e.Static("/", "static")

	// health check
	e.GET("/health-check", func(c echo.Context) error {
		return c.String(http.StatusOK, "service is healthy")
	})

	// token refresh
	e.POST("/token/refresh", authService.HandleRefreshToken)

	// user endpoints
	e.GET("/user", middleware.MiddlewareJWT(authService.HandleGetUser))
	e.POST("/user/create", middleware.MiddlewareClientCredentials(authService.HandleCreateUser))
	e.POST("/user/login", middleware.MiddlewareClientCredentials(authService.HandleLoginUser))
	e.PATCH("/user/update", middleware.MiddlewareJWT(authService.HandleUpdateUser))
	e.DELETE("/user/delete", middleware.MiddlewareJWT(authService.HandleDeleteUser))

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
