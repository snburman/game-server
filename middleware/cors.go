package middleware

import "github.com/labstack/echo/v4"

func MiddlewareCORS(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie")
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		return next(c)
	}
}
