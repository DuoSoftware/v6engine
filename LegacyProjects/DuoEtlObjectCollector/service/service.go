package service

import (
	"duov6.com/DuoEtlObjectCollector/logger"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"github.com/twinj/uuid"
	"io/ioutil"
	"net/http"
	"os"
)

func Start() {
	logger.Log("Starting Duo ETL Object Collector Endpoint....")
	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	//INSERT
	m.Post("/:namespace/:class", handleInsertRequest)
	//UPDATE
	m.Put("/:namespace/:class", handleUpdateRequest)
	//DELETE
	m.Delete("/:namespace/:class", handleDeleteRequest)

	m.RunOnAddr(":7000")
}

func handleInsertRequest(params martini.Params, res http.ResponseWriter, req *http.Request) { // res and req are injected by Martini
	logger.Log("Saving INSERT Doc :" + params["namespace"] + " class : " + params["class"])
	makeDirectories(params["namespace"], params["class"])
	rb, _ := ioutil.ReadAll(req.Body)
	SaveInsertJsons(rb, params["namespace"], params["class"])
}

func handleUpdateRequest(params martini.Params, res http.ResponseWriter, req *http.Request) { // res and req are injected by Martini
	logger.Log("Saving UPDATE Doc :" + params["namespace"] + " class : " + params["class"])
	makeDirectories(params["namespace"], params["class"])
	rb, _ := ioutil.ReadAll(req.Body)
	SaveUpdateJsons(rb, params["namespace"], params["class"])
}

func handleDeleteRequest(params martini.Params, res http.ResponseWriter, req *http.Request) { // res and req are injected by Martini
	logger.Log("Saving DELETE Doc :" + params["namespace"] + " class : " + params["class"])
	makeDirectories(params["namespace"], params["class"])
	rb, _ := ioutil.ReadAll(req.Body)
	SaveDeleteJsons(rb, params["namespace"], params["class"])
}

func getFileSavePath() (path string) {
	content, err := ioutil.ReadFile("config.config")
	if err != nil {
		fmt.Println(err.Error())
		path = ""
	} else {
		var settings map[string]interface{}
		settings = make(map[string]interface{})
		err = json.Unmarshal(content, &settings)
		if err != nil || settings["Path"] == nil {
			fmt.Println(err.Error())
			path = ""
		} else {
			path = settings["Path"].(string)
		}
	}

	return path
}

func SaveInsertJsons(Item []byte, namespace string, class string) {
	saveTempObjects(Item, 1, namespace, class)
}

func SaveUpdateJsons(Item []byte, namespace string, class string) {
	saveTempObjects(Item, 2, namespace, class)
}

func SaveDeleteJsons(Item []byte, namespace string, class string) {
	saveTempObjects(Item, 3, namespace, class)
}

func saveTempObjects(Item []byte, operation int, namespace string, class string) {
	savePath := ""
	if operation == 1 {
		savePath += (getFileSavePath() + "/" + namespace + "/" + class + "/add/new/")
	} else if operation == 2 {
		savePath += (getFileSavePath() + "/" + namespace + "/" + class + "/edit/new/")
	} else if operation == 3 {
		savePath += (getFileSavePath() + "/" + namespace + "/" + class + "/delete/new/")
	}

	err := ioutil.WriteFile((savePath + getFileName(namespace, class) + ".txt"), Item, 0666)
	if err != nil {
		fmt.Println(err.Error())
	}

}

func getFileName(namespace string, class string) string {
	return (namespace + "-" + class + "-" + uuid.NewV1().String())
}

func makeDirectories(namespace string, class string) {
	_, errr := os.Stat(getFileSavePath())
	if errr != nil {
		os.Mkdir((getFileSavePath()), 0777)
	}

	_, err := os.Stat((getFileSavePath() + "/" + namespace))
	if err != nil {
		os.Mkdir((getFileSavePath() + "/" + namespace), 0777)
		os.Mkdir((getFileSavePath() + "/" + namespace + "/" + class), 0777)
		os.Mkdir((getFileSavePath() + "/" + namespace + "/" + class + "/add"), 0777)
		os.Mkdir((getFileSavePath() + "/" + namespace + "/" + class + "/add/new"), 0777)
		os.Mkdir((getFileSavePath() + "/" + namespace + "/" + class + "/add/processing"), 0777)
		os.Mkdir((getFileSavePath() + "/" + namespace + "/" + class + "/add/completed"), 0777)
		os.Mkdir((getFileSavePath() + "/" + namespace + "/" + class + "/edit"), 0777)
		os.Mkdir((getFileSavePath() + "/" + namespace + "/" + class + "/edit/new"), 0777)
		os.Mkdir((getFileSavePath() + "/" + namespace + "/" + class + "/edit/processing"), 0777)
		os.Mkdir((getFileSavePath() + "/" + namespace + "/" + class + "/edit/completed"), 0777)
		os.Mkdir((getFileSavePath() + "/" + namespace + "/" + class + "/delete"), 0777)
		os.Mkdir((getFileSavePath() + "/" + namespace + "/" + class + "/delete/new"), 0777)
		os.Mkdir((getFileSavePath() + "/" + namespace + "/" + class + "/delete/processing"), 0777)
		os.Mkdir((getFileSavePath() + "/" + namespace + "/" + class + "/delete/completed"), 0777)
	}
}
