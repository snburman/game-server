package utils

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/snburman/game_server/config"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

func GenerateJWT(UserID string, expiry time.Duration) string {
	// set claims
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
		UserID: UserID,
	}
	// generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// encode token
	t, err := token.SignedString([]byte(config.Env().SECRET))
	if err != nil {
		panic(err)
	}
	return t
}

func ParseJWTHeader(c echo.Context) (string, error) {
	auth := c.Request().Header["Authorization"]
	if len(auth) == 0 {
		return "", errors.New("missing_authorization_header")
	}
	bearerToken := strings.Split(auth[0], " ")
	if len(bearerToken) != 2 {
		return "", errors.New("malformed_authorization_header")
	}
	return bearerToken[1], nil
}

func DecodeJWT(token string) (*JWTClaims, error) {
	t, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.Env().SECRET), nil
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	claims, ok := t.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid_jwt_claims")
	}
	if len(claims.UserID) == 0 {
		return nil, errors.New("missing_jwt_claims")
	}
	return claims, nil
}
