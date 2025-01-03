package utils

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	// password must contain one lowercase, one uppercase, one special character, and be at least 8 characters long
	errWeakPassword := errors.New("weak_password")
	if len(password) < 8 {
		return "", errWeakPassword
	}
	upper := false
	lower := false
	number := false
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsLower(c):
			lower = true
		case unicode.IsNumber(c):
			number = true
		default:
		}
	}
	if !(upper && lower && number) {
		return "", errWeakPassword
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
