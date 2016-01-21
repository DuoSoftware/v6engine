package endpoints

import (
	"duov6.com/duonotifier/client"
	"duov6.com/duonotifier/messaging"
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
	fmt.Println("DuoNotifier Listening on Port : 7000")
	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"securityToken", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	m.Post("/:namespace", handleRequest)
	m.RunOnAddr(":7000")
}

func handleRequest(params martini.Params, w http.ResponseWriter, r *http.Request) {
	namespace := params["namespace"]
	requestBody, _ := ioutil.ReadAll(r.Body)

	switch namespace {
	case "GetTemplate":
		request := getTemplateRequest(requestBody)
		response := client.GetTemplate(request)
		temp, _ := json.Marshal(response)
		fmt.Fprintf(w, "%s", string(temp))
		break
	default:
		fmt.Fprintf(w, "%s", "Method Not Found!")
		break
	}

}

func getTemplateRequest(body []byte) messaging.TemplateRequest {
	var templateRequest messaging.TemplateRequest

	err := json.Unmarshal(body, &templateRequest)
	if err != nil {
		fmt.Println(err.Error())
	}
	return templateRequest
}
