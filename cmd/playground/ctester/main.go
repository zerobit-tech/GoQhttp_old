package main

// int c = 1;
import "C"
import "fmt"

func main() {
	fmt.Println(C.c)
}
