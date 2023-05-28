package main

import (
	"fmt"

	"github.com/onlysumitg/GoQhttp/utils/stringutils"
)

func main() {

	const MySecret string = "Ang&1*~U^2^#s0^=)^^7%b34"
	fmt.Println([]byte(MySecret))
	x, _ := stringutils.Encrypt("sumit", MySecret)
	fmt.Println("1 ::  ", x)
	y, _ := stringutils.Decrypt(x, MySecret)
	fmt.Println("2 ::  ", y)

}
