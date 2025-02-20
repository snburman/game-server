package config

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Environment variables
type Vars struct {
	SERVER_URL      string
	ALLOWED_ORIGINS string
	PORT            string
	MONGO_URI       string
	SECRET          string
	CLIENT_ID       string
	CLIENT_SECRET   string
	ADMIN_ID        string
}

// Env() returns Vars struct of environment variables
func Env() Vars {
	// Load if not a test. This isn't required during testing.
	if flag.Lookup("test.v") == nil {
		err := godotenv.Load()
		if err != nil {
			log.Println("Error loading environment variables")
		}
	}

	return Vars{
		SERVER_URL:      os.Getenv("SERVER_URL"),
		ALLOWED_ORIGINS: os.Getenv("ALLOWED_ORIGINS"),
		PORT:            os.Getenv("PORT"),
		MONGO_URI:       os.Getenv("MONGO_URI"),
		SECRET:          os.Getenv("SECRET"),
		CLIENT_ID:       os.Getenv("CLIENT_ID"),
		CLIENT_SECRET:   os.Getenv("CLIENT_SECRET"),
		ADMIN_ID:        os.Getenv("ADMIN_ID"),
	}
}
