package main

import (
	"fmt"
	"reflect"
)

type myString string

type myInt int

func main() {

	x := myString("hello")

	xType := reflect.TypeOf(x)

	fmt.Println("x is ", x, xType, xType.Kind())

	y := "hellp0"
	if string(x) == y {
		fmt.Println("not same")
	}

	i := myInt(23)

	iType := reflect.TypeOf(i)

	fmt.Println("i is ", i, iType, iType.Kind())

}
