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
	isSuccess := true
	var request structs.ServiceRequest
	err := getServiceRequest(r, &request, params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		response := repositories.Execute(request)
		if response.Err != nil {
			fmt.Println(response.Err.Error())
			isSuccess = false
		} else {
			fmt.Println("Process Completed!")
			isSuccess = true
		}
	}

	if isSuccess {
		w.WriteHeader(200)

	} else {
		w.WriteHeader(500)
	}

}

func getServiceRequest(r *http.Request, request *structs.ServiceRequest, params martini.Params) (err error) {

	if r.Method != "GET" {
		rb, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		} else {
			err = json.Unmarshal(rb, &request)
			if err != nil {
				return err
			} else {
				fmt.Println(string(rb))
			}
		}
	}

	return
}
