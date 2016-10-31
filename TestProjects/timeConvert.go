package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("AAAA")

	//layout := "2014-09-12T11:45:26.371Z"
	// str := "2014-11-12T11:45:26.371Z"
	// tim, err := time.Parse(time.RFC3339, str)

	loc, _ := time.LoadLocation("UTC")
	str := "2014-11-12 11:45:26"

	tim, err := time.Parse("2006-01-02 15:04:05", str)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(tim)
		fmt.Println(tim.In(loc))

		dd := tim.Format(time.RFC3339)
		fmt.Println(dd)
	}
}
