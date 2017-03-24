package main

import (
	"fmt"
)

func main() {
	fmt.Println("Main Function")
	funcOne("Prasad", func(s bool) {
		if s {
			fmt.Println(3)
		} else {
			fmt.Println(4)
		}
	})

	fmt.Println("Hue : 0")
}

func funcOne(serverClass string, callback func(s bool)) {
	fmt.Println("Func One : " + serverClass)

	funcTwo(serverClass, func(s bool) {
		if s {
			callback(s)
		}
	})

	fmt.Println("Hue : 1")
}

func funcTwo(serverClass string, callback func(s bool)) {
	fmt.Println("Func Two : " + serverClass)

	funcThree(serverClass, func(s bool) {
		if s {
			callback(s)
		}
	})

	fmt.Println("Hue : 2")

}

func funcThree(serverClass string, callback func(s bool)) {
	fmt.Println("Func Three : " + serverClass)
	callback(true)
	fmt.Println("Hue : 3")
}
