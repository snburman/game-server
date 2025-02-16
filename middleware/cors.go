package middleware

import (
	"github.com/labstack/echo/v4"
)

func MiddlewareCORS(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// allowedOrigins := strings.Split(config.Env().ALLOWED_ORIGINS, ",")
		// origin := c.Request().Header.Get("Origin")
		// log.Println("origin", origin)
		// for _, allowedOrigin := range allowedOrigins {
		// 	if origin == allowedOrigin {
		// 		c.Response().Header().Set("Access-Control-Allow-Origin", origin)
		// 		break
		// 	}
		// }
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Client_id, Client_secret, Origin")
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		c.Response().Header().Set("Access-Control-Expose-Headers", "Connection")
		return next(c)
	}
}
