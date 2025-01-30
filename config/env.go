package config

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Environment variables
type Vars struct {
	HOST                 string
	PORT                 string
	MONGO_URI            string
	SECRET               string
	CLIENT_ID            string
	CLIENT_SECRET        string
	GOOGLE_CLIENT_ID     string
	GOOGLE_CLIENT_SECRET string
}

// Env() returns Vars struct of environment variables
func Env() Vars {
	// Load if not a test. This isn't required during testing.
	if flag.Lookup("test.v") == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading environment variables")
		}
	}

	return Vars{
		HOST:                 os.Getenv("HOST"),
		PORT:                 os.Getenv("PORT"),
		MONGO_URI:            os.Getenv("MONGO_URI"),
		SECRET:               os.Getenv("SECRET"),
		CLIENT_ID:            os.Getenv("CLIENT_ID"),
		CLIENT_SECRET:        os.Getenv("CLIENT_SECRET"),
		GOOGLE_CLIENT_ID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GOOGLE_CLIENT_SECRET: os.Getenv("GOOGLE_CLIENT_SECRET"),
	}
}
