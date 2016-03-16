package endpoints

import (
	"duov6.com/consoleworker/repositories"
	"duov6.com/consoleworker/structs"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"io/ioutil"
	"net/http"
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

	m.RunOnAddr(":7500")
}

func statusHandlder(params martini.Params, w http.ResponseWriter, r *http.Request) {
	versionData := "{\"Name\":\"Service Console Worker\", \"Status\": \"Running\"}"
	fmt.Fprintf(w, versionData)
}

func handleRequest(params martini.Params, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Post Request!")

	var requestBody structs.ServiceRequest

	if r.Method != "GET" {
		rb, rerr := ioutil.ReadAll(r.Body)
		if rerr != nil {
			fmt.Println(rerr.Error())
		} else {
			err := json.Unmarshal(rb, &requestBody)
			if err != nil {
				fmt.Println(err.Error())
				fmt.Println("Request in String : ")
				fmt.Println(string(rb))
			} else {
				response := repositories.Execute(requestBody)
				if response.Err != nil {
					fmt.Println(response.Err.Error())
				} else {
					fmt.Println("Process Completed!")
				}
			}
		}
	}

}
