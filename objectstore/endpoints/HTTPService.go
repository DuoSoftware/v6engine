package endpoints

import (
	"duov6.com/FileServer"
	FileServerMessaging "duov6.com/FileServer/messaging"
	"duov6.com/authlib"
	"duov6.com/objectstore/backup"
	"duov6.com/objectstore/configuration"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/processors"
	"duov6.com/objectstore/repositories"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"io/ioutil"
	"net/http"
	//	"runtime"
	"strings"
)

type HTTPService struct {
}

type FileData struct {
	Id       string
	FileName string
	Body     string
}

var isLoggable bool
var isJsonStack bool

func (h *HTTPService) Start(isLogEnabled bool, isJsonStackEnabled bool) {
	isLoggable = isLogEnabled
	isJsonStack = isJsonStackEnabled
	term.Write("Object Store Listening on Port : 3000", 2)
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
	//READ BY KEY
	m.Get("/:namespace/:class/:id", handleRequest)
	//READ BY KEYWORD
	m.Get("/:namespace/:class", handleRequest)
	//Get all classes
	m.Post("/:namespace", handleRequest)
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

func versionHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {
	versionData := "{\"name\": \"Objectstore\",\"version\": \"1.0.13-a\",\"Change Log\":\"Added connection pooling, PK constraint to MySQL!\",\"author\": {\"name\": \"Duo Software\",\"url\": \"http://www.duosoftware.com/\"},\"repository\": {\"type\": \"git\",\"url\": \"https://github.com/DuoSoftware/v6engine/\"}}"
	fmt.Fprintf(w, versionData)
}

func uploadHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {
	// This will upload the file as a raw file and data as record wise.
	var sendRequest = FileServerMessaging.FileRequest{}
	sendRequest.WebRequest = r
	sendRequest.WebResponse = w
	sendRequest.Parameters = make(map[string]string)
	sendRequest.Parameters = params
	headerToken := r.Header.Get("securityToken")

	if headerToken == "" {
		headerToken = "securityToken"
	}
	sendRequest.Parameters["securityToken"] = headerToken
	sendRequest.Parameters["fileContent"] = string(r.Header.Get("fileContent"))

	//Get Configuration and Read for Insert Single/Multiple Repository Selection
	configObject := configuration.ConfigurationManager{}.Get(headerToken, params["namespace"], params["class"])
	blockSize := "1000"
	for _, value := range configObject.StoreConfiguration["INSERT-MULTIPLE"] {
		if value == "ELASTIC" {
			blockSize = "100" //If Elastic is there reduce Transfer block size to 200
			break
		}
	}
	sendRequest.Parameters["BlockSize"] = blockSize

	exe := FileServer.FileManager{}
	fileResponse := exe.Store(&sendRequest)
	if fileResponse.IsSuccess == true {
		fmt.Fprintf(w, ":File uploaded successfully!")
	} else {
		fmt.Fprintf(w, "Aborted")
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
		responseMessage = getQueryResponseString("Invalid Query Request", message, false, objectRequest.MessageStack, nil)
	} else {

		dispatcher := processors.Dispatcher{}
		var repResponse repositories.RepositoryResponse = dispatcher.Dispatch(&objectRequest)
		isSuccess = repResponse.IsSuccess

		if isSuccess {
			if repResponse.Body != nil {
				responseMessage = string(repResponse.Body)
			} else {
				responseMessage = getQueryResponseString("Successfully completed request", repResponse.Message, isSuccess, objectRequest.MessageStack, repResponse.Data)
			}

		} else {
			responseMessage = getQueryResponseString("Error occured while processing", repResponse.Message, isSuccess, objectRequest.MessageStack, nil)
		}

	}

	return
}

func getQueryResponseString(mainError string, reason string, isSuccess bool, messageStack []string, Data []map[string]interface{}) string {
	response := messaging.ResponseBody{}
	response.Data = Data
	response.IsSuccess = isSuccess
	response.Message = mainError + ":" + reason
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
	sendMetaData := r.Header.Get("sendMetaData")
	headerLog := r.Header.Get("log")

	var headerOperation string
	headerMultipliciry := r.Header.Get("multiplicity")

	headerNamespace := params["namespace"]
	headerClass := params["class"]

	headerId := params["id"]
	headerKeyword := r.URL.Query().Get("keyword")

	//check if <Skip> and <Take> are specified
	//If so store them in ObjectRequest <Extras>

	if r.URL.Query().Get("skip") != "" {
		objectRequest.Extras["skip"] = r.URL.Query().Get("skip")
	}

	if r.URL.Query().Get("take") != "" {
		objectRequest.Extras["take"] = r.URL.Query().Get("take")
	}

	if r.URL.Query().Get("orderBy") != "" {
		objectRequest.Extras["orderby"] = r.URL.Query().Get("orderBy")
	}

	if r.URL.Query().Get("orderByDsc") != "" {
		objectRequest.Extras["orderbydsc"] = r.URL.Query().Get("orderByDsc")
	}

	if r.URL.Query().Get("orderby") != "" {
		objectRequest.Extras["orderby"] = r.URL.Query().Get("orderby")
	}

	if r.URL.Query().Get("fieldName") != "" {
		objectRequest.Extras["fieldName"] = r.URL.Query().Get("fieldName")
	}

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

		//isTokenValid, _ := validateSecurityToken(headerToken, headerNamespace)
		isTokenValid := true

		if isTokenValid {

			if r.Method != "GET" {
				rb, rerr := ioutil.ReadAll(r.Body)

				if rerr != nil {
					message = "Error converting request : " + rerr.Error()
					isSuccess = false
				} else {

					if isJsonStack {
						//Start writing to JsonStack
						if r.Method == "POST" {
							var temprequestBody messaging.RequestBody
							_ = json.Unmarshal(rb, &temprequestBody)
							if temprequestBody.Object != nil || temprequestBody.Objects != nil {
								backup.SaveInsertJsons(rb, headerNamespace, headerClass)
							}
						} else if r.Method == "PUT" {
							backup.SaveUpdateJsons(rb, headerNamespace, headerClass)
						} else if r.Method == "DELETE" {
							backup.SaveDeleteJsons(rb, headerNamespace, headerClass)
						}
						//Writing to JsonStack ends here
					}

					err := json.Unmarshal(rb, &requestBody)
					if err != nil {
						message = "JSON Parse error in Request : " + err.Error()
						isSuccess = false
					} else {
						//Print All Everytime. Testing for DuoAuth. Remove this after testing done.
						if isLoggable {
							fmt.Println("-----------------------------------------------------------------------------")
							fmt.Println("Primary Key : " + requestBody.Parameters.KeyProperty)
							fmt.Print("Query : ")
							fmt.Println(requestBody.Query)
							fmt.Print("Special : ")
							fmt.Println(requestBody.Special)
							fmt.Print("Single Object : ")
							fmt.Println(requestBody.Object)
							fmt.Print("Multiple Objects : ")
							fmt.Println(requestBody.Objects)
							fmt.Println("-----------------------------------------------------------------------------")
						}
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
						term.Write("Insert by POST : "+objectRequest.Body.Parameters.KeyProperty, 2)
						headerOperation = "insert"
						if len(objectRequest.Body.Object) != 0 {
							headerId = objectRequest.Body.Object[objectRequest.Body.Parameters.KeyProperty].(string)
						}
					} else if requestBody.Query.Type != "" && requestBody.Query.Type != " " {
						term.Write("Query Function Identified!", 2)
						headerOperation = "read-filter"
						canAddHeader = false
					} else if requestBody.Special.Type != "" && requestBody.Special.Type != " " {
						term.Write("Special Function Identified!", 2)
						headerOperation = "special"
						canAddHeader = false
					}

				case "PUT": //update
					if len(objectRequest.Body.Objects) != 0 {
						headerOperation = "update"
					} else {
						headerId = objectRequest.Body.Object[objectRequest.Body.Parameters.KeyProperty].(string)
						headerOperation = "update"
					}

				case "DELETE": //delete
					if len(objectRequest.Body.Objects) != 0 {
						headerOperation = "delete"
					} else {
						headerId = objectRequest.Body.Object[objectRequest.Body.Parameters.KeyProperty].(string)
						headerOperation = "delete"
					}
				}

				if len(objectRequest.Body.Objects) != 0 {
					headerMultipliciry = "multiple"
				} else if len(objectRequest.Body.Object) != 0 {
					headerMultipliciry = "single"
				}

				objectRequest.Controls = messaging.RequestControls{SecurityToken: headerToken, SendMetaData: sendMetaData, Namespace: headerNamespace, Class: headerClass, Multiplicity: headerMultipliciry, Id: headerId, Operation: headerOperation}

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

func validateSecurityToken(token string, domain string) (isValidated bool, cert authlib.AuthCertificate) {
	isValidated = true

	handler := authlib.AuthHandler{}
	cert, error := handler.GetSession(token, domain)

	if len(error) != 0 {
		isValidated = false
	}

	return
}

func checkIfFile(params martini.Params) (fileType string) {
	//Check if this a file and RETURN the file type
	var tempArray []string
	tempArray = strings.Split(params["id"], ".")
	if len(tempArray) > 1 {
		fileType = tempArray[len(tempArray)-1]
	} else {
		fileType = "NAF"
	}
	return
}
