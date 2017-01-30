package endpoints

import (
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
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
	m.Get("/:namespace/:class", handleRequest)

	m.Post("/:namespace/:class", handleRequest)
	m.Get("/:namespace/:class", handleRequest)

	m.Post("/:namespace/:class/:id", handleRequest)
	m.Get("/:namespace/:class/:id", handleRequest)

	m.RunOnAddr(":8500")
}

func statusHandlder(params martini.Params, w http.ResponseWriter, r *http.Request) {
	versionData := "{\"Name\":\"REST Tester\", \"Status\": \"Running\"}"
	fmt.Fprintf(w, versionData)
}

func handleRequest(params martini.Params, w http.ResponseWriter, r *http.Request) {
	fileHandlers = make(map[string]*os.File)
	var err error
	rb, err := ioutil.ReadAll(r.Body)

	jsonData := make(map[string]interface{})

	json.Unmarshal(rb, &jsonData)

	PrintLogs(jsonData)

	if err == nil {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(500)
	}

}

func PrintLogs(input map[string]interface{}) {
	fileName := input["InSessionID"].(string) + ".txt"
	Body := "This is A Sample Body"

	for x := 0; x < 30; x++ {
		PublishLog(fileName, Body)
		//time.Sleep(200 * time.Millisecond)
	}

}

var fileHandlers map[string]*os.File
var fileHandlerLock = sync.RWMutex{}

func GetFileHandler(index string) (conn *os.File) {
	fileHandlerLock.RLock()
	defer fileHandlerLock.RUnlock()
	conn = fileHandlers[index]
	return
}

func SetFileHandler(index string, conn *os.File) {
	fileHandlerLock.Lock()
	defer fileHandlerLock.Unlock()
	fileHandlers[index] = conn
}

func PublishLog(fileName string, Body string) {
	var slash = ""
	if runtime.GOOS == "windows" {
		slash = "\\"
	} else {
		slash = "/"
	}
	var logfolderName = "Logs"

	// check if the publishing activity folder exists, if not create one.

	logrootpath := logfolderName
	_, activityRooterr := os.Stat(logrootpath)
	if activityRooterr != nil {
		// create folder in the given path and permissions
		os.Mkdir(logrootpath, 0777)
	}

	if runtime.GOOS == "linux" {

		date := string(time.Now().Local().Format("2006-01-02 @ 15:04:05"))
		_, _ = exec.Command("sh", "-c", "echo "+date+"    "+Body+" >> "+logfolderName+slash+fileName).Output()
		//_, _ = exec.Command("sh", "-c", "echo "+Body+" >> "+fileName).Output()

	} else {

		var ff *os.File
		var err error
		name := logfolderName + slash + fileName
		if GetFileHandler(name) != nil {
			ff = GetFileHandler(name)
		} else {
			ff, err = os.OpenFile(name, os.O_APPEND, 0666)
			if err != nil {
				ff, err = os.Create(name)
				ff, err = os.OpenFile(name, os.O_APPEND, 0666)
			}
			SetFileHandler(name, ff)
		}

		_, err = ff.Write([]byte(string(time.Now().Local().Format("2006-01-02 @ 15:04:05")) + "  " + Body + "  " + "\r\n"))
		if err != nil {
			fmt.Println(err.Error() + "asdf")
		}
	}
}
