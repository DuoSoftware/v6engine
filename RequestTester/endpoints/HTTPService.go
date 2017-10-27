package endpoints

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"io/ioutil"
	"net/http"
	"time"
)

type HTTPService struct {
}

func (h *HTTPService) Start() {
	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"securityToken", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	m.Get("/", statusHandlder)
	m.Post("/:namespace/:class", handleRequest)
	m.Get("/:namespace/:class", handleRequest)

	m.Post("/:namespace/:class", handleRequest)
	m.Get("/:namespace/:class", handleRequest)

	m.Post("/:namespace/:class/:id", handleRequest)
	m.Get("/:namespace/:class/:id", handleRequest)

	m.RunOnAddr(":8500")
}

func statusHandlder(params martini.Params, w http.ResponseWriter, r *http.Request) {
	versionData := "{\"Name\":\"REST Tester\", \"Status\": \"Running\"}"
	fmt.Fprintf(w, versionData)
}

func handleRequest(params martini.Params, w http.ResponseWriter, r *http.Request) {
	isSuccess := true

	color.Green("---------------- Request @ " + time.Now().UTC().String() + "-------------------")
	color.Yellow("Request Type --> " + r.Method)
	fmt.Println()
	color.Yellow("Headers --> ")
	for key, value := range r.Header {
		fmt.Print(key + " : ")
		fmt.Println(value)
	}
	fmt.Println()
	color.Yellow("URL Parameters --> ")
	for key, value := range r.URL.Query() {
		fmt.Print(key + " : ")
		fmt.Println(value)
	}
	fmt.Println()
	color.Yellow("Body --> ")
	rb, _ := ioutil.ReadAll(r.Body)
	fmt.Println(string(rb))
	fmt.Println()
	fmt.Println()

	if isSuccess {
		w.WriteHeader(200)
		fmt.Fprintf(w, "%s", "{}")
	} else {
		w.WriteHeader(500)
	}

}
