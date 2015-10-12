package endpoint

import (
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
)

type HTTPService struct {
}

type layoutRequest struct {
	HTML       string
	Parameters map[string]string
}

func (h *HTTPService) Start() {
	fmt.Println("Layout Engine Listening on Port : 7000")
	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"securityToken", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	m.Post("/:namespace/:class", handleRequest)

	m.RunOnAddr(":7000")
}

func handleRequest(params martini.Params, res http.ResponseWriter, req *http.Request) { // res and req are injected by Martini

	requestBody, _ := ioutil.ReadAll(req.Body)

	var layoutReq layoutRequest
	_ = json.Unmarshal(requestBody, &layoutReq)

	fullName := params["namespace"] + "-" + params["class"] + "-" + layoutReq.HTML + ".html"
	content, _ := ioutil.ReadFile(fullName)
	fileContent := string(content)
	lines := strings.Split(fileContent, "\n")

	var variableMap map[int]string
	variableMap = make(map[int]string)

	index := 0
	for _, singleLine := range lines {
		tempWords := strings.Split(singleLine, " ")
		for _, word := range tempWords {
			if strings.Index(word, "@") != (-1) {
				varList := strings.Split(word, "@")
				variableMap[index] = varList[1]
				index++
			}
		}
	}

	for _, value := range variableMap {
		fileContent = strings.Replace(fileContent, ("@" + value + "@"), value, -1)
	}

	var replceMap map[string]string
	replceMap = make(map[string]string)
	replceMap = layoutReq.Parameters

	for variable, value := range replceMap {
		fileContent = strings.Replace(fileContent, variable, value, -1)
	}

	returnHTML := params["namespace"] + "-" + params["class"] + "-" + layoutReq.HTML + "converted.html"
	returnPDF := params["namespace"] + "-" + params["class"] + "-" + layoutReq.HTML + ".pdf"

	ioutil.WriteFile(returnHTML, []byte(fileContent), 0666)
	_, _ = exec.Command("sh", "-c", ("wkhtmltopdf " + returnHTML + " " + returnPDF)).Output()

}
