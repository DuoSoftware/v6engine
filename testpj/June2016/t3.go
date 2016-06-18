package main

import (
	"fmt"
)

func main() {
	pp := make([]map[string]interface{}, 0)

	gg := make(map[string]interface{})
	gg["1"] = "a"
	gg1 := make(map[string]interface{})
	gg1["1"] = "b"
	gg2 := make(map[string]interface{})
	gg2["1"] = "c"
	gg3 := make(map[string]interface{})
	gg3["1"] = "d"

	pp = append(pp, gg)
	pp = append(pp, gg1)
	pp = append(pp, gg2)
	pp = append(pp, gg3)

	fmt.Println(pp)

}
