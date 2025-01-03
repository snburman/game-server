package config

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Type Vars
//
// Environment variables
type Vars struct {
	MONGO_URI            string
	SECRET               string
	ClientID             string
	ClientSecret         string
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
		MONGO_URI:            os.Getenv("MONGO_URI"),
		SECRET:               os.Getenv("SECRET"),
		ClientID:             os.Getenv("CLIENT_ID"),
		ClientSecret:         os.Getenv("CLIENT_SECRET"),
		GOOGLE_CLIENT_ID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GOOGLE_CLIENT_SECRET: os.Getenv("GOOGLE_CLIENT_SECRET"),
	}
}
