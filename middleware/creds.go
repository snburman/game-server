package middleware

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/config"
	"github.com/snburman/game_server/errors"
)

type ClientCredentialsDTO struct {
	ClientID     string      `json:"client_id"`
	ClientSecret string      `json:"client_secret"`
	Data         interface{} `json:"data"`
}

type ClientDataContext struct {
	echo.Context
	Data interface{}
}

func MiddlewareClientCredentials(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// bind credentials
		var creds ClientCredentialsDTO
		err := c.Bind(&creds)
		if err != nil {
			return c.JSON(
				http.StatusUnauthorized,
				errors.ServerError(errors.ErrInvalidCredentials).JSON(),
			)
		}
		// compare credentials
		if creds.ClientID != config.Env().CLIENT_ID || creds.ClientSecret != config.Env().CLIENT_SECRET {
			return c.JSON(
				http.StatusUnauthorized,
				errors.ServerError(errors.ErrInvalidCredentials).JSON(),
			)
		}
		// construct context
		ctx := ClientDataContext{
			Context: c,
			Data:    creds.Data,
		}

		return next(ctx)
	}
}

func UnmarshalClientDataContext[T any](c echo.Context) (T, error) {
	ctx, ok := c.(ClientDataContext)
	data := *new(T)
	if !ok {
		log.Println("missing_client_context")
		return data, c.NoContent(http.StatusInternalServerError)
	}
	b, err := json.Marshal(ctx.Data)
	if err != nil {
		log.Println(err)
		return data, c.NoContent(http.StatusInternalServerError)
	}

	err = json.Unmarshal(b, &data)
	if err != nil {
		log.Println(err)
		return data, c.NoContent(http.StatusInternalServerError)
	}

	return data, nil
}

func MiddleWareClientHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		clientID := c.Request().Header.Get("CLIENT_ID")
		clientSecret := c.Request().Header.Get("CLIENT_SECRET")

		if clientID != config.Env().CLIENT_ID || clientSecret != config.Env().CLIENT_SECRET {
			return c.NoContent(http.StatusUnauthorized)
		}
		return next(c)
	}
}
