package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT         string
	MODE         string
	TOKEN        string
	APP_URL      string
	USER_SERVICE_URL string
	LOAN_SERVICE_URL string
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
	Env.APP_URL = os.Getenv("APP_URL")
	Env.MODE = os.Getenv("MODE")
	Env.TOKEN = os.Getenv("TOKEN")
	Env.USER_SERVICE_URL = os.Getenv("USER_SERVICE_URL")
	Env.LOAN_SERVICE_URL = os.Getenv("LOAN_SERVICE_URL")
}
