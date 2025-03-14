package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/snburman/game-server/errors"
	"github.com/snburman/game-server/utils"
)

type JWTContext struct {
	echo.Context
	*utils.JWTClaims
}

func MiddlewareJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := utils.ParseJWTHeader(c)
		if token == "" || err != nil {
			return c.JSON(
				http.StatusUnauthorized,
				errors.AuthenticationError(errors.ErrInvalidJWT).JSON(),
			)
		}
		claims, err := utils.DecodeJWT(token)
		if err != nil || claims.UserID == "" {
			return c.JSON(
				http.StatusUnauthorized,
				errors.AuthenticationError(errors.ErrInvalidJWT).JSON(),
			)
		}
		ctx := JWTContext{
			Context:   c,
			JWTClaims: claims,
		}
		return next(ctx)
	}
}
