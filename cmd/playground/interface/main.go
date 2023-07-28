package main

import "fmt"

// https://golangbot.com/interfaces-part-1/
//https://golangbot.com/interfaces-part-2/

// -----------------------------------------------------------
//
// -----------------------------------------------------------
type myinterface interface {
	// ~ int | ~ int64    >>> https://stackoverflow.com/questions/71073365/interface-contains-type-constraints-cannot-use-interface-in-conversion
	ToString() (returnval string)

	MakeDouble(input int) (doubled int)
}

// -----------------------------------------------------------
//
// -----------------------------------------------------------
type MyType struct {
	i int
}

func (m *MyType) ToString() string {
	return fmt.Sprintf("%d", m.i)
}

func (m *MyType) MakeDouble(input int) int {
	return input * 2
}

// -----------------------------------------------------------
//
// -----------------------------------------------------------
type MyType2 int

func (m MyType2) ToString() string {
	return fmt.Sprintf("%d", m)
}

func (m MyType2) MakeDouble(input int) int {
	return input * 2
}

func describe(w myinterface) {
	fmt.Printf("Interface type %T value %v\n", w, w)
	fmt.Println("x3 is nil", w == nil)

}

// -----------------------------------------------------------
//
// -----------------------------------------------------------
func main() {
	var x myinterface = MyType2(23)
	var x2 myinterface = &MyType{i: 223}

	fmt.Println(x.ToString(), x.MakeDouble(30))
	fmt.Println(x2.ToString(), x2.MakeDouble(30))

	var myType *MyType = nil
	var x3 myinterface = myType
	// myType is nil
	// but x3 is not nil
	//    base interhave is tuple of (type, value)
	//  interface is nil when both its type and value are nil
	// in above case even the value is nil
	//    type is not nill
	// check the describe methods above

	fmt.Println("myType is", myType)

	describe(x3)

	//Type assertion

	// Type assertion is used to extract the underlying value of the interface.

	//Type switch

	//A type switch is used to compare the concrete type of an interface against multiple types specified in various case statements. It is similar to switch case. The only difference being the cases specify types and not values as in normal switch.
}

func findType(i interface{}) {
	switch i.(type) {
	case string:
		fmt.Printf("I am a string and my value is %s\n", i.(string))
	case int:
		fmt.Printf("I am an int and my value is %d\n", i.(int))
	default:
		fmt.Printf("Unknown type\n")
	}
}
