package endpoints

import (
	"duov6.com/FileServer"
	FileServerMessaging "duov6.com/FileServer/messaging"
	"duov6.com/authlib"
	"duov6.com/cebadapter"
	"duov6.com/common"
	"duov6.com/objectstore/JSON_Purifier"
	"duov6.com/objectstore/backup"
	"duov6.com/objectstore/cache"
	"duov6.com/objectstore/configuration"
	"duov6.com/objectstore/keygenerator"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/processors"
	"duov6.com/objectstore/repositories"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
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

var isLoggable bool
var isJsonStack bool
var isFlusherActivated bool

func (h *HTTPService) Start() {
	if !common.VerifyGlobalConfig() {
		//GetConfigs from REST...
		if status := cebadapter.GetGlobalConfigFromREST("StoreConfig"); !status {
			fmt.Println("Error retrieving configurations from CEB... Exiting...")
			os.Exit(1)
		}
	}

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

	//------- Utility End Points -------------

	//Get All Error Post Logs
	m.Get("/ErrorLogs", logHandler)

	//Sync Increment Keys with DomainClassAttributes
	m.Get("/SyncRedisKeys", syncHandler)

	//Flush Cache
	m.Get("/ClearCache", cacheHandler)

	//Get Store Configurations
	m.Get("/config", getConfigHandler)

	//View All Logs
	m.Get("/ViewLogs")

	//Enable or Disable Terminal View For Request Body
	m.Get("/ToggleLogs", viewLogHandler)

	//Enable or Disable logging Requests to Disk
	m.Get("/ToggleStack", jsonStackHandler)

	//5.1 silverlight access
	// m.Get("/crossdomain.xml", Crossdomain)
	// m.Get("/clientaccesspolicy.xml", Clientaccesspolicy)
	m.Run()

}

func startKeyFlusher(request *messaging.ObjectRequest) {
	if !isFlusherActivated {
		isFlusherActivated = true
		if repositories.CheckRedisAvailability(request) {
			go keygenerator.UpdateCountsToDB()
		}
	}
}

func viewLogHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {
	msg := term.ToggleConfig()
	if strings.Contains(msg, "Enabled") {
		isLoggable = true
		martini.IsOSLogEnabled = true
	} else {
		isLoggable = false
		martini.IsOSLogEnabled = false
	}
	fmt.Fprintf(w, msg)
}

func jsonStackHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {
	msg := ""
	if isJsonStack {
		isJsonStack = false
		msg = "Disabled Writing to JSON Stack!"
	} else {
		isJsonStack = true
		msg = "Enabled Writing to JSON Stack!"
	}
	fmt.Fprintf(w, msg)
}

func syncHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {
	keygenerator.UpdateKeysInDB()
	fmt.Fprintf(w, "Syncing Redis Keys!")
}

func logHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {
	objectRequest := messaging.ObjectRequest{}
	paramMap := make(map[string]interface{})
	objectRequest.Extras = paramMap
	_, _ = getObjectRequest(r, &objectRequest, params)

	message := ""
	if params["class"] == "" {
		if CheckRedisAvailability(&objectRequest) {
			keyArray := cache.GetKeyListPattern(&objectRequest, "*", cache.Log)
			keyArrayInBytes, _ := json.Marshal(keyArray)
			message = string(keyArrayInBytes)
		} else {
			message = "Error! REDIS not configured in this server!"
		}
	} else {
		message = string(cache.GetKeyValue(&objectRequest, params["class"], cache.Log))
	}

	fmt.Fprintf(w, message)
}

func getConfigHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {
	configAll := cebadapter.GetGlobalConfig("StoreConfig")
	byteArray, _ := json.Marshal(configAll)
	fmt.Fprintf(w, string(byteArray))
}

func cacheHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {
	params["namespace"] = "ignore"
	params["class"] = "ignore"
	objectRequest := messaging.ObjectRequest{}
	paramMap := make(map[string]interface{})
	objectRequest.Extras = paramMap
	_, _ = getObjectRequest(r, &objectRequest, params)
	message := ""

	if CheckRedisAvailability(&objectRequest) {
		cache.FlushCache(&objectRequest)
		message = "REDIS cache was successfully cleared!"
	} else {
		mongoRepo := repositories.Create("MONGO")
		mongoRepo.ClearCache(&objectRequest)

		cassandraRepo := repositories.Create("CASSANDRA")
		cassandraRepo.ClearCache(&objectRequest)

		hiveRepo := repositories.Create("HIVE")
		hiveRepo.ClearCache(&objectRequest)

		postgresRepo := repositories.Create("POSTGRES")
		postgresRepo.ClearCache(&objectRequest)

		mssqlRepo := repositories.Create("MSSQL")
		mssqlRepo.ClearCache(&objectRequest)

		cloudsqlRepo := repositories.Create("CLOUDSQL")
		cloudsqlRepo.ClearCache(&objectRequest)

		message = "REDIS not configured in this server. Available In-Memory Data structures cleared!"
	}
	fmt.Fprintf(w, message)
}

// func Crossdomain(params martini.Params, w http.ResponseWriter, r *http.Request) {

// 	file, err := os.Open("crossdomain.xml")
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	} else {
// 		data, err := ioutil.ReadAll(file)
// 		if err != nil {
// 			fmt.Println(err.Error())
// 		} else {
// 			w.Write(data)
// 		}
// 	}
// 	defer file.Close()
// }

// func Clientaccesspolicy(params martini.Params, w http.ResponseWriter, r *http.Request) {

// 	file, err := os.Open("clientaccesspolicy.xml")
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	} else {
// 		data, err := ioutil.ReadAll(file)
// 		if err != nil {
// 			fmt.Println(err.Error())
// 		} else {
// 			w.Write(data)
// 		}
// 	}
// 	defer file.Close()
// }

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
		fmt.Fprintf(w, " : File uploaded successfully!")
	} else {
		fmt.Fprintf(w, "Aborted")
	}
}

func handleRequest(params martini.Params, res http.ResponseWriter, req *http.Request) { // res and req are injected by Martini

	if params["namespace"] == "ErrorLogs" {
		logHandler(params, res, req)
		return
	}

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
		responseMessage = getQueryResponseString("Invalid Query Request", message, false, objectRequest.MessageStack, nil, messaging.TransactionResponse{})
	} else {
		startKeyFlusher(&objectRequest)
		dispatcher := processors.Dispatcher{}
		var repResponse repositories.RepositoryResponse = dispatcher.Dispatch(&objectRequest)
		isSuccess = repResponse.IsSuccess

		if isSuccess {
			if repResponse.Body != nil {
				responseMessage = string(repResponse.Body)
			} else {
				responseMessage = getQueryResponseString("Successfully completed request", repResponse.Message, isSuccess, objectRequest.MessageStack, repResponse.Data, repResponse.Transaction)
			}

		} else {
			responseMessage = getQueryResponseString("Error occured while processing", repResponse.Message, isSuccess, objectRequest.MessageStack, nil, messaging.TransactionResponse{})
		}

	}

	return
}

func getQueryResponseString(mainError string, reason string, isSuccess bool, messageStack []string, Data []map[string]interface{}, Transaction messaging.TransactionResponse) string {
	response := messaging.ResponseBody{}
	response.Data = Data
	response.IsSuccess = isSuccess
	response.Transaction = Transaction
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

	if r.Header.Get("searchGlobalNamespace") != "" {
		objectRequest.Extras["searchGlobalNamespace"] = r.Header.Get("searchGlobalNamespace")
	}

	if r.Header.Get("timezone") != "" {
		objectRequest.Extras["timezone"] = r.Header.Get("timezone")
	}

	if r.URL.Query().Get("timezone") != "" {
		objectRequest.Extras["timezone"] = r.URL.Query().Get("timezone")
	}

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

	if r.URL.Query().Get("securityToken") != "" {
		headerToken = r.URL.Query().Get("securityToken")
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
		var initialSlice []string
		initialSlice = make([]string, 0)
		objectRequest.MessageStack = initialSlice
	}

	var requestBody messaging.RequestBody

	if isSuccess {

		//isTokenValid, _ := validateSecurityToken(headerToken, headerNamespace)
		isTokenValid := true

		if isTokenValid {

			if r.Method != "GET" {
				rb, rerr := ioutil.ReadAll(r.Body)
				//Clean JSON with escape characters
				rb = JSON_Purifier.Purify(rb)

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
						color.Red("---------------------------- ERROR REQUEST BODY -----------------------------")
						color.Red(string(rb))
						color.Red("-----------------------------------------------------------------------------")
					} else {
						if isLoggable {
							color.Cyan("---------------------------- REQUEST BODY -----------------------------------")
							color.Cyan(string(rb))
							color.Cyan("-----------------------------------------------------------------------------")
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
						//term.Write("Insert by POST : "+objectRequest.Body.Parameters.KeyProperty, 2)
						headerOperation = "insert"
						if len(objectRequest.Body.Object) != 0 {
							headerId = objectRequest.Body.Object[objectRequest.Body.Parameters.KeyProperty].(string)
						}
					} else if requestBody.Query.Type != "" && requestBody.Query.Type != " " {
						//term.Write("Query Function Identified!", 2)
						headerOperation = "read-filter"
						canAddHeader = false
					} else if requestBody.Special.Type != "" && requestBody.Special.Type != " " {
						//term.Write("Special Function Identified!", 2)
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

func CheckRedisAvailability(request *messaging.ObjectRequest) (status bool) {
	status = true
	if request.Configuration.ServerConfiguration["REDIS"] == nil {
		status = false
	}
	return
}

//------------------ Version Management -----------------------

func versionHandler(params martini.Params, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, GetVersion())
}
