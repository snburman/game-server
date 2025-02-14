package middleware

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

func MiddlewareCORS(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Println("Request from: ", c.Request().RemoteAddr)
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Client_id, Client_secret")
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		c.Response().Header().Set("Access-Control-Expose-Headers", "Connection")
		return next(c)
	}
}
