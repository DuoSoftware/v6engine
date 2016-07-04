package main

import (
	"duov6.com/common"
	"fmt"
)

func main() {
	params := make(map[string]string)
	params["securityToken"] = "123"
	params["log"] = "log"
	fmt.Println(common.HTTP_GET("http://obj.duoworld.com:3000/com.duosoftware.auth/users", params, true))
}
