package main

import "fmt"
import "time"

func main() {
	c := time.Tick(5 * time.Second)
	for now := range c {
		fmt.Println(now)
	}
}
