package main

import (
	"fmt"
	"time"
)

func main() {
	objectTime, _ := time.Parse(time.RFC3339, "2017-01-20T06:00:31Z")
	fmt.Println(objectTime)
	timeDifference := time.Now().UTC().Sub(objectTime)
	fmt.Println(int(timeDifference.Hours()))
}
