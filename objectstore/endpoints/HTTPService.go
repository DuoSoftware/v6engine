package endpoints

import (
	"duov6.com/authlib"
	"duov6.com/objectstore/configuration"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/processors"
	"duov6.com/objectstore/repositories"
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
	fmt.Println("Object Store Listening on Port : 3000")
	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"securityToken", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	//READ BY KEY
	m.Get("/:namespace/:class/:id", handleRequest)
	//READ BY KEYWORD
	m.Get("/:namespace/:class", handleRequest)
	//READ ADVANCED, INSERT
	m.Post("/:namespace/:class", handleRequest)

	//UPDATE
	m.Put("/:namespace/:class", handleRequest)
	//DELETE
	m.Delete("/:namespace/:class", handleRequest)

	m.Run()
}

func handleRequest(params martini.Params, res http.ResponseWriter, req *http.Request) { // res and req are injected by Martini

	responseMessage, isSuccess := dispatchRequest(req, params)

	if isSuccess {
		res.WriteHeader(200)
	} else {
		res.WriteHeader(500)
	}

	fmt.Fprintf(res, "%s", responseMessage)
}

func (h *HTTPService) Stop() {
}

func dispatchRequest(r *http.Request, params martini.Params) (responseMessage string, isSuccess bool) { //result is JSON

	objectRequest := messaging.ObjectRequest{}

	paramMap := make(map[string]interface{})
	objectRequest.Extras = paramMap

	message, isSuccess := getObjectRequest(r, &objectRequest, params)

	if isSuccess == false {
		responseMessage = getQueryResponseString("Invalid Query Request", message, false, objectRequest.MessageStack)
	} else {

		dispatcher := processors.Dispatcher{}
		var repResponse repositories.RepositoryResponse = dispatcher.Dispatch(&objectRequest)
		isSuccess = repResponse.IsSuccess

		if isSuccess {
			if repResponse.Body != nil {
				responseMessage = string(repResponse.Body)
			} else {
				responseMessage = getQueryResponseString("Successfully completed request", repResponse.Message, isSuccess, objectRequest.MessageStack)
			}

		} else {
			responseMessage = getQueryResponseString("Error occured while processing", repResponse.Message, isSuccess, objectRequest.MessageStack)
		}

	}

	return
}

func getQueryResponseString(mainError string, reason string, isSuccess bool, messageStack []string) string {
	response := messaging.ResponseBody{}
	response.IsSuccess = isSuccess
	response.Message = mainError + " : " + reason
	if messageStack != nil {
		response.Stack = messageStack
	}

	result, err := json.Marshal(&response)

	if err == nil {
		return string(result)
	} else {
		return "Invalid Query"
	}
}

func getObjectRequest(r *http.Request, objectRequest *messaging.ObjectRequest, params martini.Params) (message string, isSuccess bool) {

	missingFields := ""
	isSuccess = true

	headerToken := r.Header.Get("securityToken")
	headerLog := r.Header.Get("log")

	var headerOperation string
	headerMultipliciry := r.Header.Get("multiplicity")

	headerNamespace := params["namespace"]
	headerClass := params["class"]

	headerId := params["id"]
	headerKeyword := r.URL.Query().Get("keyword")

	if len(headerToken) == 0 {
		isSuccess = false
		missingFields = missingFields + "securityToken"
	}

	if len(headerLog) != 0 {
		objectRequest.IsLogEnabled = true
		var initialSlice []string
		initialSlice = make([]string, 0)
		objectRequest.MessageStack = initialSlice
	} else {
		objectRequest.IsLogEnabled = false
	}

	var requestBody messaging.RequestBody

	if isSuccess {

		isTokenValid, _ := validateSecurityToken(headerToken)

		if isTokenValid {

			if r.Method != "GET" {
				rb, rerr := ioutil.ReadAll(r.Body)

				if rerr != nil {
					message = "Error converting request : " + rerr.Error()
					isSuccess = false
				} else {

					err := json.Unmarshal(rb, &requestBody)

					if err != nil {
						message = "JSON Parse error in Request : " + err.Error()
						isSuccess = false
					} else {
						objectRequest.Body = requestBody
					}
				}
			}

			if isSuccess {

				canAddHeader := true
				switch r.Method {
				case "GET": //read keyword, and unique key
					if len(headerId) != 0 {
						headerOperation = "read-key"
					} else if len(headerKeyword) != 0 {
						objectRequest.Body = messaging.RequestBody{}
						objectRequest.Body.Query = messaging.Query{Parameters: headerKeyword}
						headerOperation = "read-keyword"
					} else if len(headerNamespace) != 0 && len(headerClass) != 0 {
						headerOperation = "read-all"
					}
					canAddHeader = false
				case "POST": //read query, read special, insert
					if len(requestBody.Object) != 0 || len(requestBody.Objects) != 0 {
						fmt.Println("Inset by POST : " + objectRequest.Body.Parameters.KeyProperty)
						headerOperation = "insert"
						if len(objectRequest.Body.Object) != 0 {
							headerId = objectRequest.Body.Object[objectRequest.Body.Parameters.KeyProperty].(string)
						}
					} else if &requestBody.Query != nil {
						headerOperation = "read-filter"
						canAddHeader = false
					}

				case "PUT": //update
					headerId = objectRequest.Body.Object[objectRequest.Body.Parameters.KeyProperty].(string)
					headerOperation = "update"

				case "DELETE": //delete
					headerId = objectRequest.Body.Object[objectRequest.Body.Parameters.KeyProperty].(string)
					headerOperation = "delete"
				}

				if len(objectRequest.Body.Objects) != 0 {
					headerMultipliciry = "multiple"
				} else if len(objectRequest.Body.Object) != 0 {
					headerMultipliciry = "single"
				}

				objectRequest.Controls = messaging.RequestControls{SecurityToken: headerToken, Namespace: headerNamespace, Class: headerClass, Multiplicity: headerMultipliciry, Id: headerId, Operation: headerOperation}

				configObject := configuration.ConfigurationManager{}.Get(headerToken, headerNamespace, headerClass)
				objectRequest.Configuration = configObject

				if canAddHeader {
					repositories.FillControlHeaders(objectRequest)
				}
			}

		} else {
			isSuccess = false
			message = "Access token not validated." + missingFields
		}
	} else {
		message = "Missing attributes in request header : " + missingFields
	}

	return
}

func validateSecurityToken(token string) (isValidated bool, cert authlib.AuthCertificate) {
	isValidated = true

	handler := authlib.AuthHandler{}
	cert, error := handler.GetSession(token)

	if len(error) != 0 {
		isValidated = false
	}

	return
}
