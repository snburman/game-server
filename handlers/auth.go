package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	gorillaSessions "github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/config"
	"github.com/snburman/game_server/db"
	"github.com/stytchauth/stytch-go/v15/stytch/consumer/magiclinks"
	emailML "github.com/stytchauth/stytch-go/v15/stytch/consumer/magiclinks/email"
	"github.com/stytchauth/stytch-go/v15/stytch/consumer/passwords"
	"github.com/stytchauth/stytch-go/v15/stytch/consumer/sessions"
	"github.com/stytchauth/stytch-go/v15/stytch/consumer/stytchapi"
	"github.com/stytchauth/stytch-go/v15/stytch/consumer/users"
)

const LOGIN_FORM = "static/login.html"

type AuthService struct {
	client *stytchapi.API
	store  *gorillaSessions.CookieStore
}

func NewAuthService(projectId, secret string) *AuthService {
	client, err := stytchapi.NewClient(projectId, secret)
	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}

	return &AuthService{
		client: client,
		store:  gorillaSessions.NewCookieStore([]byte(config.Env().SECRET)),
	}
}

func (a *AuthService) HandleCreateUser(c echo.Context) error {
	var user db.User
	err := c.Bind(&user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, nil)
	}

	if user.Email == "" || user.Password == "" || user.UserName == "" {
		return c.JSON(http.StatusBadRequest, ErrMissingParams.JSON())
	}

	user, err = a.createUserAuth(c, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	profile := db.NewUserProfile(user)
	err = db.CreateUserProfile(db.MongoDB, profile)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrCreatingProfile.JSON())
	}

	return c.JSON(http.StatusCreated, profile)
}

func (a *AuthService) HandleLoginUser(c echo.Context) error {
	var user db.User
	err := c.Bind(&user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrMissingParams.JSON())
	}

	if user.Email == "" || user.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrMissingParams.JSON())
	}

	params := &passwords.AuthenticateParams{
		Email:    user.Email,
		Password: user.Password,
	}

	resp, err := a.client.Passwords.Authenticate(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrInvalidCredentials.JSON())
	}

	profile, err := db.GetUserProfileByID(db.MongoDB, resp.UserID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, profile)
}

func (a *AuthService) HandleDeleteUser(c echo.Context) error {
	var user db.User
	err := c.Bind(&user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrMissingParams.JSON())
	}
	if user.ID == "" {
		return c.JSON(http.StatusBadRequest, ErrMissingParams.JSON())
	}

	err = a.deleteUserAuth(c, user)
	if err != nil {
		return err
	}
	count, err := db.DeleteUserProfile(db.MongoDB, user.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, struct {
		Deleted int `json:"deleted"`
	}{
		Deleted: count,
	})
}

func (a *AuthService) HandlePasswordReset(c echo.Context) error {
	return nil
}

func (a *AuthService) createUserAuth(c echo.Context, u db.User) (db.User, error) {
	params := &passwords.CreateParams{
		Email:    u.Email,
		Password: u.Password,
	}

	// Create user on auth platform
	resp, err := a.client.Passwords.Create(c.Request().Context(), params)
	if err != nil {
		return *new(db.User), ErrInvalidCredentials
	}
	u.ID = resp.UserID
	return u, nil
}

func (a *AuthService) deleteUserAuth(c echo.Context, u db.User) error {
	// delete from auth platform
	params := &users.DeleteParams{
		UserID: u.ID,
	}
	_, err := a.client.Users.Delete(c.Request().Context(), params)
	if err != nil {
		return err
	}
	return nil
}

/////////////////////////////////////////////////////
// Future email link implementation
/////////////////////////////////////////////////////

func (a *AuthService) HandleMagicLink(c echo.Context) error {
	if err := c.Request().ParseForm(); err != nil {
		log.Printf("Error parsing form: %v\n", err)
		return err
	}

	email := c.Request().Form.Get("email")
	if email == "" {
		return c.JSON(http.StatusBadRequest, nil)
	}

	resp, err := a.client.MagicLinks.Email.LoginOrCreate(
		c.Request().Context(),
		&emailML.LoginOrCreateParams{
			Email: email,
		})
	if err != nil {
		log.Printf("Error sending email: %v\n", err)
		return err
	}

	fmt.Println(resp)

	return c.JSON(http.StatusOK, nil)
}

func (a *AuthService) HandleAuthenticate(c echo.Context) error {
	tokenType := c.Request().URL.Query().Get("stytch_token_type")
	token := c.Request().URL.Query().Get("token")

	if tokenType != "magic_links" {
		return c.JSON(http.StatusInternalServerError, nil)
	}

	res, err := a.client.MagicLinks.Authenticate(
		c.Request().Context(),
		&magiclinks.AuthenticateParams{
			Token:                  token,
			SessionDurationMinutes: 60 * 24,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	session, err := a.store.Get(c.Request(), "stytch_session")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	session.Values["token"] = res.SessionToken
	session.Save(c.Request(), c.Response())

	return c.JSON(http.StatusOK, nil)
}

func (a *AuthService) HandleAuthIndex(c echo.Context) error {
	user := a.getAuthenticatedUser(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, nil)
	}
	return c.JSON(http.StatusAccepted, user)
}

func (a *AuthService) getAuthenticatedUser(c echo.Context) *users.User {
	res, req := c.Response(), c.Request()
	session, err := a.store.Get(req, "stytch_session")
	if err != nil || session == nil {
		return nil
	}

	token, ok := session.Values["token"].(string)
	if !ok || token == "" {
		return nil
	}

	resp, err := a.client.Sessions.Authenticate(
		context.Background(),
		&sessions.AuthenticateParams{
			SessionToken: token,
		})
	if err != nil {
		delete(session.Values, "token")
		session.Save(req, res)
		return nil
	}
	session.Values["token"] = resp.SessionToken
	session.Save(req, res)

	return &resp.User
}

func MiddlewareCORS(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return next(c)
	}
}
