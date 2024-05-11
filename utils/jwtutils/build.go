package jwtutils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zerobit-tech/GoQhttp/env"
)

func getSecret() []byte {
	hmacSampleSecret := env.GetEnvVariable("JWTKEY", "6c4XmaXkfx6CL_968b7t3oYUyUZ8X86c4XmaXkfx6CL_968b7t3oYUyUZ8X8")
	return []byte(hmacSampleSecret)
}

// ----------------------------------------------------------
//
// ----------------------------------------------------------
func Get(claims map[string]any) (string, error) {

	//GetNotBefore
	claims["nbf"] = time.Now().UTC().Unix()
	claims["iss"] = "QHTTP"

	// expiry

	_, found := claims["exp"]
	if !found {
		claims["exp"] = time.Now().UTC().Add(300 * time.Minute).Unix()
	}

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	//token := jwt.New(jwt.SigningMethodHS256)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(getSecret())

	return tokenString, err
}

// ----------------------------------------------------------
//
// ----------------------------------------------------------
func Parse(tokenString string) (map[string]any, error) {

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return getSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}

}
