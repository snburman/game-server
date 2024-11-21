package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/snburman/magic_game_server/db"
	"github.com/snburman/magic_game_server/utils"
)

const LOGIN_FORM = "static/login.html"

func HandleLogin(c echo.Context) error {
	// token := c.Request().Header.Get("Authorization")
	// if token != "" {
	// 	// Check if the token is valid

	// }
	return c.JSON(http.StatusOK, nil)

}

func HandleAuthenticate(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest.JSON())
	}

	// Get username and password from request body
	var creds db.User
	err := json.NewDecoder(c.Request().Body).Decode(&creds)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest.JSON())
	}
	username = creds.Username
	password = creds.Password

	// Check if user exists
	user, err := db.GetUser(db.MongoDB.Client, username)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrInvalidCredentials.JSON())
	}

	// Check if password is valid
	if !utils.CheckPasswordHash(password, user.Password) {
		return c.JSON(http.StatusUnauthorized, ErrInvalidCredentials.JSON())
	}

	return HandleCreateSession(c)
}

func HandleRegister(c echo.Context) error {
	// Get the username and password from the request
	username := c.FormValue("username")
	password := c.FormValue("password")

	// Check if user already exists
	user, err := db.GetUser(db.MongoDB.Client, username)
	if err == nil && user.Username != "" {
		return c.JSON(http.StatusConflict, ErrUserExists.JSON())
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrCreatingUser.JSON())
	}

	res, err := db.CreateUser(db.MongoDB.Client, db.User{
		Username: username,
		Password: hash,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrCreatingUser.JSON())
	}

	return c.JSON(http.StatusCreated, res)
}

func HandleLogout(c echo.Context) error {
	return nil
}

func MiddlewareCORS(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return next(c)
	}
}
