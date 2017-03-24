package endpoints

import (
	"duov6.com/common"
	"duov6.com/duonotifier/client"
	"duov6.com/duonotifier/messaging"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
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

	m.Get("/", versionHandler)
	m.Post("/:namespace", handleRequest)
	m.RunOnAddr(":7000")
}

func versionHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {
	cpuUsage := strconv.Itoa(int(common.GetProcessorUsage()))
	cpuCount := strconv.Itoa(runtime.NumCPU())
	//versionDaata := "{\"Name\": \"Objectstore\",\"Version\": \"1.4.4-a\",\"Change Log\":\"Fixed certain alter table issues.\",\"Author\": {\"Name\": \"Duo Software\",\"URL\": \"http://www.duosoftware.com/\"},\"Repository\": {\"Type\": \"git\",\"URL\": \"https://github.com/DuoSoftware/v6engine/\"},\"System Usage\": {\"CPU\": \" " + cpuUsage + " (percentage)\",\"CPU Cores\": \"" + cpuCount + "\"}}"
	versionData := make(map[string]interface{})
	versionData["API Name"] = "Duo Notifier"
	versionData["API Version"] = "6.1.00"

	versionData["Change Log"] = [...]string{
		"Started new versioning with 6.1.00",
		"Added agent.config to reflect localhost if agent.config not found",
	}

	gitMap := make(map[string]string)
	gitMap["Type"] = "git"
	gitMap["URL"] = "https://github.com/DuoSoftware/v6engine/"
	versionData["Repository"] = gitMap

	statMap := make(map[string]string)
	statMap["CPU"] = cpuUsage + " (percentage)"
	statMap["CPU Cores"] = cpuCount
	versionData["System Usage"] = statMap

	authorMap := make(map[string]string)
	authorMap["Name"] = "Duo Software Pvt Ltd"
	authorMap["URL"] = "http://www.duosoftware.com/"
	versionData["Project Author"] = authorMap

	byteArray, _ := json.Marshal(versionData)

	fmt.Fprintf(w, string(byteArray))
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
