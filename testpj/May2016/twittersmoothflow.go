package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"io/ioutil"
	"net/http"
)

func main() {
	Start()
}

func Start() {

	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"securityToken", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	//Read Version
	m.Get("/", versionHandler)

	m.Get("/:namespace", handleRequest)
	m.Post("/:namespace", handleRequest)
	m.Put("/:namespace", handleRequest)

	m.RunOnAddr(":3333")
}

func versionHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {
	versionData := "{\"name\": \"TWITTER CALL BAKC TESTERRRRRRRRRRRRRRRR\",\"version\": \"1.2.6-a\",\"Change Log\":\"Experimental type conversion!\",\"author\": {\"name\": \"Duo Software\",\"url\": \"http://www.duosoftware.com/\"},\"repository\": {\"type\": \"git\",\"url\": \"https://github.com/DuoSoftware/v6engine/\"}}"
	fmt.Fprintf(w, versionData)
}

func handleRequest(params martini.Params, res http.ResponseWriter, req *http.Request) { // res and req are injected by Martini
	fmt.Println("METHOD : " + req.Method)

	rb, rerr := ioutil.ReadAll(req.Body)

	if rerr != nil {
		fmt.Println(rerr.Error())
	} else {
		fmt.Println(string(rb))
	}

	res.WriteHeader(200)
	responseMessage := "Successful AF!"
	fmt.Fprintf(res, "%s", responseMessage)
}
