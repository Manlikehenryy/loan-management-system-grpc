package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT               string
	MONGO_DB_URI       string
	MODE               string
	TOKEN              string
	USER_SERVICE_URL   string
	WALLET_SERVICE_URL string
}

var Env *Config

func init() {

	Env = &Config{}

	if os.Getenv("MODE") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}

	}

	Env.PORT = os.Getenv("PORT")
	Env.MONGO_DB_URI = os.Getenv("MONGO_DB_URI")
	Env.MODE = os.Getenv("MODE")
	Env.TOKEN = os.Getenv("TOKEN")
	Env.USER_SERVICE_URL = os.Getenv("USER_SERVICE_URL")
	Env.WALLET_SERVICE_URL = os.Getenv("WALLET_SERVICE_URL")
}
