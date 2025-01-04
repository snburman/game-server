package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/db"
	"github.com/snburman/game_server/errors"
	"github.com/snburman/game_server/middleware"
	"github.com/snburman/game_server/utils"
)

type AuthService struct {
	store *sessions.CookieStore
}

type AuthResponse struct {
	errors.ServerError `json:"error,omitempty"`
	Token              string `json:"token"`
	RefreshToken       string `json:"refresh_token"`
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (a *AuthService) HandleRefreshToken(c echo.Context) error {
	rt, err := middleware.UnmarshalClientDataContext[string](c)
	// var params struct {
	// 	RefreshToken string `json:"refresh_token"`
	// 	ClientID     string `json:"client_id"`
	// 	ClientSecret string `json:"client_secret"`
	// }
	// err := c.Bind(&params)
	if err != nil {
		log.Println("missing_refresh_token")
		return c.NoContent(http.StatusUnauthorized)
	}
	claims, err := utils.DecodeJWT(rt)
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
	claims, ok := c.(middleware.JWTContext)
	if !ok {
		return c.JSON(http.StatusUnauthorized, AuthResponse{
			ServerError: errors.ErrInvalidJWT,
		})
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
			ServerError: errors.ErrUserBanned,
		})
	}
	return c.JSON(http.StatusOK, user)
}

func (a *AuthService) HandleCreateUser(c echo.Context) error {
	// get user from context
	u, err := middleware.UnmarshalClientDataContext[db.User](c)
	if err != nil {
		return err
	}
	// check if user exists
	user, err := db.GetUserByUserName(db.MongoDB, u.UserName)
	if err == nil && user.UserName == u.UserName {
		return c.JSON(http.StatusInternalServerError, AuthResponse{
			ServerError: errors.ErrUserExists,
		})
	}
	// create user
	id, err := db.CreateUser(db.MongoDB, u)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AuthResponse{
			ServerError: errors.ServerError(err.Error()),
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
	// get user data from context
	u, err := middleware.UnmarshalClientDataContext[db.User](c)
	if err != nil {
		return err
	}
	if u.UserName == "" || u.Password == "" {
		return c.JSON(http.StatusUnauthorized, AuthResponse{
			ServerError: errors.ErrMissingParams,
		})
	}
	// get user from db
	user, err := db.GetUserByUserName(db.MongoDB, u.UserName)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, AuthResponse{
			ServerError: errors.ErrInvalidCredentials,
		})
	}
	// validate password
	passwordValid := utils.CheckPasswordHash(u.Password, user.Password)
	if !passwordValid {
		return c.JSON(http.StatusForbidden, AuthResponse{
			ServerError: errors.ErrInvalidCredentials,
		})
	}
	// reject if user banned
	if user.Banned {
		return c.JSON(http.StatusForbidden, AuthResponse{
			ServerError: errors.ErrUserBanned,
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

func (a *AuthService) HandleUpdateUser(c echo.Context) error {
	var user db.User
	err := c.Bind(&user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.ErrMissingParams.JSON())
	}
	err = db.UpdateUser(db.MongoDB, user)
	if err != nil {
		if err.Error() == errors.ErrWeakPassword.Error() {
			return c.JSON(http.StatusBadRequest, errors.ErrWeakPassword.JSON())
		}
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, errors.ErrUpdatingUser.JSON())
	}
	return c.NoContent(http.StatusAccepted)
}

func (a *AuthService) HandleDeleteUser(c echo.Context) error {
	claims, ok := c.(middleware.JWTContext)
	if !ok {
		return c.JSON(http.StatusUnauthorized, errors.ErrInvalidJWT.JSON())
	}
	if claims.UserID == "" {
		return c.JSON(http.StatusBadRequest, errors.ErrMissingParams.JSON())
	}
	count, err := db.DeleteUser(db.MongoDB, claims.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.ServerError(err.Error()).JSON())
	}

	return c.JSON(http.StatusAccepted, struct {
		Deleted int `json:"deleted"`
	}{
		Deleted: count,
	})
}
