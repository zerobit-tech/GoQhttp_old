package env

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// ----------------------------------------------------------------
//
// ----------------------------------------------------------------
func GetEnvVariable(key string, defaultVal string) string {

	// load .env file
	//// It's important to note that it WILL NOT OVERRIDE an env variable that already exists - consider the .env file to set dev vars or sensible defaults.
	err := godotenv.Overload("env/.env")

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

// ----------------------------------------------------------------
//
// ----------------------------------------------------------------
func IsInDebugMode() bool {
	debugS := strings.TrimSpace(strings.ToUpper(GetEnvVariable("DEBUG", "FALSE")))

	if debugS == "TRUE" || debugS == "YES" || debugS == "Y" {
		return true
	}

	return false
}

// ----------------------------------------------------------------
//
// ----------------------------------------------------------------
func GetServerPassword(serverName string) string {
	serverName = strings.ToUpper(strings.TrimSpace(serverName))
	envVarName := fmt.Sprintf("%s_PASSWORD", serverName)
	pwd := GetEnvVariable(envVarName, "")
	return pwd
}

// ----------------------------------------------------------------
//
// ----------------------------------------------------------------
func GetServerUserName(serverName string) string {
	serverName = strings.ToUpper(strings.TrimSpace(serverName))
	envVarName := fmt.Sprintf("%s_USER", serverName)
	u := GetEnvVariable(envVarName, "")
	return u
}

// ----------------------------------------------------------------
//
// ----------------------------------------------------------------
func AllowHtmlTemplates() bool {
	debugS := strings.TrimSpace(strings.ToUpper(GetEnvVariable("ALLOWHTMLTEMPLATES", "FALSE")))

	if debugS == "TRUE" || debugS == "YES" || debugS == "Y" {
		return true
	}

	return false
}

// ----------------------------------------------------------------
//
// ----------------------------------------------------------------
func UserSessionDuration() time.Duration {
	i := strings.TrimSpace(strings.ToUpper(GetEnvVariable("USERSESSIONTIMEOUT", "120")))

	iDurataion, err := strconv.Atoi(i)
	if err != nil {
		iDurataion = 120
	}

	return time.Duration(iDurataion * int(time.Minute))
}
