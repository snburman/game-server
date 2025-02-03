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
	e.POST("/token/refresh", middleware.MiddlewareClientCredentials(authService.HandleRefreshToken))

	// user endpoints
	e.GET("/user", middleware.MiddlewareJWT(authService.HandleGetUser))
	e.POST("/user/create", middleware.MiddlewareClientCredentials(authService.HandleCreateUser))
	e.POST("/user/login", middleware.MiddlewareClientCredentials(authService.HandleLoginUser))
	e.PATCH("/user/update", middleware.MiddlewareJWT(authService.HandleUpdateUser))
	e.DELETE("/user/delete", middleware.MiddlewareJWT(authService.HandleDeleteUser))

	// game
	//
	e.GET("/game/client/connect", middleware.MiddlewareWebSocket(handlers.HandleClientConnect))
	e.GET("/game/client", handlers.HandleGetGame)
	// initiated by game wasm to retrieve map by ID
	e.GET("/game/wasm/map/:id", middleware.MiddleWareClientHeaders(handlers.HandleGetMapByID))

	// assets
	//
	// all assets
	e.GET("/assets", middleware.MiddlewareJWT(handlers.HandleGetAssets))
	// assets by player
	e.GET("/assets/player", middleware.MiddlewareJWT(handlers.HandleGetPlayerAssets))
	e.POST("/assets/player", middleware.MiddlewareJWT(handlers.HandleCreatePlayerAsset))
	e.PATCH("/assets/player", middleware.MiddlewareJWT(handlers.HandleUpdatePlayerAsset))
	e.DELETE("/assets/player", middleware.MiddlewareJWT(handlers.HandleDeletePlayerAsset))

	// maps
	e.GET("/maps/:id", middleware.MiddlewareJWT(handlers.HandleGetMapByID))
	e.POST("/maps", middleware.MiddlewareJWT(handlers.HandleCreateMap))
	e.GET("/maps/player", middleware.MiddlewareJWT(handlers.HandleGetPlayerMaps))
	e.PATCH("/maps", middleware.MiddlewareJWT(handlers.HandleUpdateMap))
	e.DELETE("/maps/:id", middleware.MiddlewareJWT(handlers.HandleDeleteMap))

	// database
	db.NewMongoDriver()

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = ":9191"
	} else {
		PORT = ":" + PORT
	}
	e.Logger.Fatal(e.Start(PORT))
}
