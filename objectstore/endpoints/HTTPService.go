package endpoints

import (
	"duov6.com/authlib"
	"duov6.com/common"
	"duov6.com/objectstore/client"
	"duov6.com/objectstore/configuration"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/processors"
	"duov6.com/objectstore/repositories"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"github.com/toqueteos/webbrowser"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type HTTPService struct {
}

type FileData struct {
	Id       string
	FileName string
	Body     string
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

	//FILE RECIEVER
	m.Post("/:namespace/:class/:id", uploadHandler)
	//UPDATE
	m.Put("/:namespace/:class", handleRequest)
	//DELETE
	m.Delete("/:namespace/:class", handleRequest)

	m.Run()
}

func uploadHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {

	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("file")

	if err != nil {
		fmt.Print(w)
		fmt.Println(err.Error())
		return
	}

	out, err := os.Create(header.Filename)
	if err != nil {
		fmt.Print(w)
		fmt.Println(", Unable to create the file for writing. Check your write access privilege")
		return
	}

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Print(w)
		fmt.Println(err.Error())
	}

	file2, err2 := ioutil.ReadFile(header.Filename)

	if err2 != nil {
		panic(err2)
	}

	convertedBody := string(file2[:])
	base64Body := common.EncodeToBase64(convertedBody)

	obj := FileData{}
	obj.Id = params["id"]
	obj.FileName = header.Filename
	obj.Body = base64Body

	headerToken := r.Header.Get("securityToken")

	client.Go(headerToken, params["namespace"], params["class"]).StoreObject().WithKeyField("Id").AndStoreOne(obj).FileOk()

	fmt.Fprintf(w, "File uploaded successfully : ")
	fmt.Fprintf(w, header.Filename)

	//close the files
	err = out.Close()
	err = file.Close()

	if err != nil {
		panic(err)
	}

	//remove the temporary stored file from the disk
	err2 = os.Remove(header.Filename)

	if err2 != nil {
		fmt.Println(err2)
	}
}

func handleRequest(params martini.Params, res http.ResponseWriter, req *http.Request) { // res and req are injected by Martini

	// Start setting up Content-Types
	if checkIfFile(params) == "NAF" {
		// NAF = Not A File.
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	} else if checkIfFile(params) == "txt" {
		res.Header().Set("Content-Type", "text/txt")
	} else if checkIfFile(params) == "docx" {
		res.Header().Set("Content-Type", "document/word")
	} else if checkIfFile(params) == "xlsx" {
		res.Header().Set("Content-Type", "document/excel")
	} else if checkIfFile(params) == "pptx" {
		res.Header().Set("Content-Type", "document/powerpoint")
	} else if checkIfFile(params) == "png" {
		res.Header().Set("Content-Type", "image/png")
	} else if checkIfFile(params) == "jpg" {
		res.Header().Set("Content-Type", "image/jpg")
	} else if checkIfFile(params) == "gif" {
		res.Header().Set("Content-Type", "image/gif")
	} else if checkIfFile(params) == "wav" {
		res.Header().Set("Content-Type", "audio/wav")
	} else if checkIfFile(params) == "mp3" {
		res.Header().Set("Content-Type", "audio/mp3")
	} else if checkIfFile(params) == "wmv" {
		res.Header().Set("Content-Type", "audio/wmv")
	} else {
		res.Header().Set("Content-Type", "text/other")
	}
	// End setting up Content-Types

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

				file := FileData{}

				json.Unmarshal(repResponse.Body, &file)

				if file.FileName == "" {
					fmt.Println("This is Not A File : Executing Standard Proceedure")
				} else {
					fmt.Println("This is A File : Executing Get Requested File Proceedure")
					pathToStore := "D:/FileServer/"      //Server Path
					pathToRedirect := "ftp://127.0.0.1/" //Server IP
					GetRequestedFile(file, pathToStore, pathToRedirect)
				}

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

	isValidated = true

	return
}

func checkIfFile(params martini.Params) (fileType string) {

	var tempArray []string
	tempArray = strings.Split(params["id"], ".")
	if len(tempArray) > 1 {
		fileType = tempArray[len(tempArray)-1]
	} else {
		fileType = "NAF"
	}
	return
}

func GetRequestedFile(file FileData, pathToStore string, pathToRedirect string) {
	temp := common.DecodeFromBase64(file.Body)
	ioutil.WriteFile("D:/FileServer/"+file.FileName, []byte(temp), 0666)
	webbrowser.Open("ftp://127.0.0.1/" + file.FileName)
}
