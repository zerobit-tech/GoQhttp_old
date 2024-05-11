package main

import (
	"fmt"
	"time"

	"github.com/zerobit-tech/GoQhttp/utils/jwtutils"
)

func main() {

	claims := map[string]any{"user": "sumit"}

	jwtString, err := jwtutils.Get(claims)
	if err != nil {
		fmt.Println("jwt err", err)

	}
	jwtutils.Parse(jwtString)

	fmt.Println("================ waiting ===========")
	time.Sleep(3 * time.Minute)

	jwtutils.Parse(jwtString)

}
