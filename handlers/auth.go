package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/db"
	"github.com/snburman/game_server/utils"
)

type AuthService struct {
	store *sessions.CookieStore
}

type AuthResponse struct {
	ServerError  `json:"error,omitempty"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (a *AuthService) HandleRefreshToken(c echo.Context) error {
	var params struct {
		RefreshToken string `json:"refresh_token"`
	}
	err := c.Bind(&params)
	if err != nil {
		log.Println("missing_refresh_token")
		return c.NoContent(http.StatusUnauthorized)
	}
	claims, err := utils.DecodeJWT(params.RefreshToken)
	if err != nil || claims.UserID == "" {
		log.Println("bad_refresh_token")
		return c.NoContent(http.StatusUnauthorized)
	}
	user, err := db.GetUserByID(db.MongoDB, claims.UserID)
	if err != nil {
		log.Println("user_not_found")
		return c.NoContent(http.StatusUnauthorized)
	}
	if user.Banned {
		log.Println("user_banned")
		return c.NoContent(http.StatusUnauthorized)
	}

	// generate token response
	token := utils.GenerateJWT(user.ID.Hex(), time.Minute*30)
	refreshToken := utils.GenerateJWT(user.ID.Hex(), time.Hour*7*24)
	res := AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}
	return c.JSON(http.StatusAccepted, res)
}

func (a *AuthService) HandleGetUser(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, AuthResponse{
			ServerError: ErrMissingParams,
		})
	}

	claims, err := utils.DecodeJWT(token)
	if err != nil || claims.UserID == "" {
		log.Println("bad_refresh_token")
		return c.NoContent(http.StatusUnauthorized)
	}
	user, err := db.GetUserByID(db.MongoDB, claims.UserID)
	if err != nil {
		log.Println("user_not_found")
		return c.NoContent(http.StatusUnauthorized)
	}
	user.Password = ""
	if user.Banned {
		log.Println("user_banned")
		return c.JSON(http.StatusForbidden, AuthResponse{
			ServerError: ErrUserBanned,
		})
	}

	return c.JSON(http.StatusOK, user)
}

func (a *AuthService) HandleCreateUser(c echo.Context) error {
	var u db.User
	err := c.Bind(&u)
	if err != nil || u.Password == "" || u.UserName == "" {
		return c.JSON(http.StatusBadRequest, AuthResponse{
			ServerError: ServerError(err.Error()),
		})
	}
	// check if user exists
	user, err := db.GetUserByUserName(db.MongoDB, u.UserName)
	if err == nil && user.UserName == u.UserName {
		return c.JSON(http.StatusInternalServerError, AuthResponse{
			ServerError: ServerError("user_exists"),
		})
	}

	// create user
	id, err := db.CreateUser(db.MongoDB, u)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AuthResponse{
			ServerError: ServerError(err.Error()),
		})
	}

	// generate token response
	token := utils.GenerateJWT(id.Hex(), time.Minute*30)
	refreshToken := utils.GenerateJWT(id.Hex(), time.Hour*7*24)
	res := AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}

	return c.JSON(http.StatusCreated, res)
}

func (a *AuthService) HandleLoginUser(c echo.Context) error {
	var u db.User
	err := c.Bind(&u)
	if err != nil || u.UserName == "" || u.Password == "" {
		return c.JSON(http.StatusBadRequest, AuthResponse{
			ServerError: ErrMissingParams,
		})
	}

	user, err := db.GetUserByUserName(db.MongoDB, u.UserName)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, AuthResponse{
			ServerError: ErrInvalidCredentials,
		})
	}

	passwordValid := utils.CheckPasswordHash(u.Password, user.Password)
	if !passwordValid {
		return c.JSON(http.StatusForbidden, AuthResponse{
			ServerError: ErrInvalidCredentials,
		})
	}

	if user.Banned {
		return c.JSON(http.StatusForbidden, AuthResponse{
			ServerError: ErrUserBanned,
		})
	}

	// generate token response
	token := utils.GenerateJWT(user.ID.Hex(), time.Minute*30)
	refreshToken := utils.GenerateJWT(user.ID.Hex(), time.Hour*7*24)
	res := AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}

	return c.JSON(http.StatusOK, res)
}

func (a *AuthService) HandleDeleteUser(c echo.Context) error {
	var user db.User
	err := c.Bind(&user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrMissingParams.JSON())
	}
	if user.ID.Hex() == "" {
		return c.JSON(http.StatusBadRequest, ErrMissingParams.JSON())
	}

	count, err := db.DeleteUser(db.MongoDB, user.ID.Hex())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, struct {
		Deleted int `json:"deleted"`
	}{
		Deleted: count,
	})
}

func (a *AuthService) HandleUpdateUser(c echo.Context) error {
	// var user db.User
	// err := c.Bind(&user)
	return nil
}

func (a *AuthService) HandlePasswordReset(c echo.Context) error {
	return nil
}

func MiddlewareCORS(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie")
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		return next(c)
	}
}
