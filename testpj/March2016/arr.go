package main

import (
	"fmt"
)

func main() {
	var values []string

	if values == nil {
		fmt.Println("nil array")
	}

	fmt.Println(len(values))

	values = append(values, "1")
	fmt.Println("-----------------")
	if values == nil {
		fmt.Println("nil array")
	}

	fmt.Println(len(values))

	fmt.Println("************")

	values = values[:0]
	if values == nil {
		fmt.Println("nil array")
	}

	fmt.Println(len(values))
	values = append(values, "1")
	fmt.Println("-----------------")
	if values == nil {
		fmt.Println("nil array")
	}

	fmt.Println(len(values))
}
