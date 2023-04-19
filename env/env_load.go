package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetEnvVariable(key string, defaultVal string) string {

	// load .env file
	err := godotenv.Load("env/.env")

	if err != nil {
		log.Println("Error loading .env file", err.Error())
	}

	val := os.Getenv(key)
	log.Println("Got ", key, val)
	if val == "" {
		val = defaultVal
	}

	return val
}
