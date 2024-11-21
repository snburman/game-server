package handlers

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var sessionManager *scs.SessionManager

func init() {
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.SameSite = http.SameSiteDefaultMode
}

func HandleGetSession(c echo.Context) error {
	value := sessionManager.Get(c.Request().Context(), "userId")
	token := sessionManager.Token(c.Request().Context())

	notFound := c.JSON(http.StatusUnauthorized, map[string]string{
		"message": "No session found",
	})
	if value == nil {
		return notFound
	}
	userId, ok := value.(string)
	if !ok || userId == "" {
		return notFound
	}
	return c.JSON(200, map[string]string{
		"userId": userId,
		"token":  token,
	})
}

func HandleCreateSession(c echo.Context) error {
	userId := uuid.New().String()
	sessionManager.Put(c.Request().Context(), "userId", userId)

	return c.JSON(http.StatusOK, map[string]string{
		"userId": userId,
	})
}

func SessionMiddleWare(next http.Handler) http.Handler {
	return sessionManager.LoadAndSave(next)
}

func HandleFindSession(c echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	b, ok, err := sessionManager.Store.Find(token)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Error finding session",
		})
	}
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"message": "No session found",
		})
	}
	if b == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"message": "No session found",
		})
	}

	return c.JSON(http.StatusOK, b)
}
