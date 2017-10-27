package endpoints

import (
	"duov6.com/cebadapter"
	"duov6.com/common"
	"duov6.com/duonotifier/client"
	"duov6.com/duonotifier/messaging"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"io/ioutil"
	"net/http"
	"os"
)

type HTTPService struct {
}

func (h *HTTPService) Start() {
	if !common.VerifyGlobalConfig() {
		//GetConfigs from REST...
		if status := cebadapter.GetGlobalConfigFromREST("StoreConfig"); !status {
			fmt.Println("Error retrieving configurations from CEB... Exiting...")
			os.Exit(1)
		}
	}

	fmt.Println("DuoNotifier Listening on Port : 7000")
	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"securityToken", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	m.Get("/", versionHandler)

	//Get Store Configurations
	m.Get("/config", getConfigHandler)

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

	fmt.Println("--------------------------------------------")
	fmt.Println("Request in String : ")
	fmt.Println(string(body))
	fmt.Println("Request in Map : ")
	err := json.Unmarshal(body, &templateRequest)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(templateRequest)
	}
	fmt.Println("--------------------------------------------")
	return templateRequest
}

func versionHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, GetVersion())
}

func getConfigHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {
	configAll := cebadapter.GetGlobalConfig("StoreConfig")
	byteArray, _ := json.Marshal(configAll)
	fmt.Fprintf(w, string(byteArray))
}
