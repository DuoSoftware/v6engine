package main

import "fmt"

import "reflect"

func main() {
	gg := make(map[string]interface{})

	gg["1"] = "fdsafad"

	a, b := gg["1"]

	fmt.Println(a)
	fmt.Println(b)

	fmt.Println(reflect.TypeOf(a))
	fmt.Println(reflect.TypeOf(b))

}
