package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/snburman/game-server/db"
	"github.com/snburman/game-server/handlers"
	"github.com/snburman/game-server/middleware"
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
	e.POST("/token/refresh", middleware.MiddleWareClientHeaders(middleware.MiddlewareClientDTO(authService.HandleRefreshToken)))

	// user endpoints
	e.GET("/user", middleware.MiddlewareJWT(authService.HandleGetUser))
	e.POST("/user/create", middleware.MiddleWareClientHeaders(middleware.MiddlewareClientDTO(authService.HandleCreateUser)))
	e.POST("/user/login", middleware.MiddleWareClientHeaders(middleware.MiddlewareClientDTO(authService.HandleLoginUser)))
	e.PATCH("/user/update", middleware.MiddlewareJWT(authService.HandleUpdateUser))
	e.DELETE("/user/delete", middleware.MiddlewareJWT(authService.HandleDeleteUser))

	// game endpoints
	//
	// websocket connection
	e.GET("/game/ws/:userID", handlers.HandleGameWebsocket)
	// client
	e.GET("/game/client", handlers.HandleGetGame)
	// wasm
	e.GET("/game/wasm/map", middleware.MiddleWareClientHeaders(handlers.HandleGetMapByID))
	e.GET("/game/wasm/map/ids", middleware.MiddleWareClientHeaders(handlers.HandleGetAllMapsByIDs))
	e.GET("/game/wasm/map/primary/:userID", middleware.MiddleWareClientHeaders(handlers.HandleGetPlayerPrimaryMap))
	e.GET("/game/wasm/player/default", middleware.MiddleWareClientHeaders(handlers.HandleGetDefaultPlayerCharacter))

	// assets
	//
	// all assets
	e.GET("/assets", middleware.MiddlewareJWT(handlers.HandleGetAssets))
	// assets by player
	e.GET("/assets/player", middleware.MiddlewareJWT(handlers.HandleGetPlayerAssets))
	// default player character
	e.POST("/assets/player", middleware.MiddlewareJWT(handlers.HandleCreatePlayerAsset))
	e.PATCH("/assets/player", middleware.MiddlewareJWT(handlers.HandleUpdatePlayerAsset))
	e.DELETE("/assets/player", middleware.MiddlewareJWT(handlers.HandleDeletePlayerAsset))

	// maps
	e.GET("/maps", middleware.MiddlewareJWT(handlers.HandleGetAllMaps))
	e.POST("/maps", middleware.MiddlewareJWT(handlers.HandleCreateMap))
	e.PATCH("/maps", middleware.MiddlewareJWT(handlers.HandleUpdateMap))
	e.GET("/maps/player", middleware.MiddlewareJWT(handlers.HandleGetPlayerMaps))
	e.GET("/maps/:id", middleware.MiddlewareJWT(handlers.HandleGetMapByID))
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
