package env

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func GetEnvVariable(key string, defaultVal string) string {

	// load .env file
	err := godotenv.Load("env/.env")

	if err != nil {
		log.Println("Error loading .env file", err.Error())
	}

	val := os.Getenv(key)
	//log.Println("Got ", key, val)
	if val == "" {
		val = defaultVal
	}

	return val
}

func IsInDebugMode() bool {
	debugS := strings.TrimSpace(strings.ToUpper(GetEnvVariable("DEBUG", "FALSE")))

	if debugS == "TRUE" || debugS == "YES" || debugS == "Y" {
		return true
	}

	return false
}
