package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	fmt.Println("START")

	file, err := ioutil.ReadFile("claims.xlsx")
	if err != nil {
		fmt.Print()
	}

	fmt.Println("END")
}
