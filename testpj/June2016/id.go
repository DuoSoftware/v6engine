package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	// pp := "900"

	// vv, err := strconv.Atoi(pp)

	// if err != nil {
	// 	fmt.Println(err.Error())
	// } else {
	// 	fmt.Println(vv)
	// }

	youString := "12IN655V00026"
	youStringLowered := strings.ToLower(youString)
	fmt.Println(youString)

	isIndexFound := false
	index := 0

	for x := 0; x < len(youString); x++ {
		_, err := strconv.Atoi(string(youString[x]))

		if err == nil {
			if !isIndexFound {
				match, _ := regexp.MatchString("([a-z]+)", youStringLowered[x:])
				if !match {
					index = x
					isIndexFound = true
				}
			}
		}
	}

	fmt.Print("Index : ")
	fmt.Println(index)

	prefix := youString[:index]
	valueInString := youString[index:]

	fmt.Println(prefix)
	fmt.Println(valueInString)

}
