package main

import (
	"duov6.com/DuoEtlService/configuration"
	"duov6.com/DuoEtlService/logger"
	"duov6.com/DuoEtlService/messaging"
	"duov6.com/DuoEtlService/repositories"
	"duov6.com/DuoEtlService/fileprocessor"
	"duov6.com/cebadapter"
	"fmt"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	draw()

	//Loading CEB
	forever := make(chan bool)
	cebadapter.Attach("DuoEtl", func(s bool) {
		cebadapter.GetLatestGlobalConfig("DuoEtl", func(data []interface{}) {
			logger.Log("Duo ETL Service Configuration Successfully Loaded...")
			if data != nil {
				fmt.Println(data)
				forever <- false
			}
		})
		logger.Log("Successfully registered in CEB")
	})

	<-forever

	var config = configuration.ConfigurationManager{}.Get()
	var etlconfig = configuration.ETLConfiguration{}
	etlconfig = config

	var Request *messaging.ETLRequest
	Request = &messaging.ETLRequest{}
	Request.Configuration = etlconfig
	
	logger.Log("Waiting for JSON Objects......")
	
	for true{
	if checkForNewFiles(Request.Configuration.DataPath){
		executeTask(Request)
	}
		time.Sleep(60 * time.Second)
	}
	//go delaySecond(300, Request)
	//select {}
}

func executeTask(request *messaging.ETLRequest) {
	logger.Log("Executing @ " + getTime())
	repositories.Dispatch(request)
}
/*
func delaySecond(n time.Duration, request *messaging.ETLRequest) {
	for _ = range time.Tick(n * time.Second) {
		executeTask(request)
	}
} 
*/
func getTime() (retTime string) {
	currentTime := time.Now().Local()
	year := strconv.Itoa(currentTime.Year())
	month := strconv.Itoa(int(currentTime.Month()))
	day := strconv.Itoa(currentTime.Day())
	hour := strconv.Itoa(currentTime.Hour())
	minute := strconv.Itoa(currentTime.Minute())
	second := strconv.Itoa(currentTime.Second())

	retTime = (year + "-" + month + "-" + day + " ... " + hour + ":" + minute + ":" + second)

	return
}

func checkForNewFiles(objectsPath string) (status bool){
	status = false
	classPaths := fileprocessor.GetClassPaths(objectsPath)
	for _, classpath := range classPaths{
		addFiles,_ := filepath.Glob(classpath +"/add/new/"+ "*.txt")
		if len(addFiles) > 0{
			status = true
			return
		}
		editFiles,_ := filepath.Glob(classpath +"/edit/new/"+ "*.txt")
		if len(editFiles) > 0{
			status = true
			return
		}
		deleteFiles,_ := filepath.Glob(classpath +"/delete/new/"+ "*.txt")
		if len(deleteFiles) > 0{
			status = true
			return
		}
	}
	return 
}

func draw() {
	logger.Log("Starting Duo ETL Service.........")
	logger.Log("\n")
	logger.Log("\n")
	logger.Log("______             _____ _____ _           ____ ")
	logger.Log("|  _  \\           |  ___|_   _| |         / ___|")
	logger.Log("| | | |_   _  ___ | |__   | | | |  __   _/ /___ ")
	logger.Log("| | | | | | |/ _ \\|  __|  | | | |  \\ \\ / / ___ \\")
	logger.Log("| |/ /| |_| | (_) | |___  | | | |___\\ V /| \\_/ |")
	logger.Log("|___/  \\__,_|\\___/\\____/  \\_/ \\_____/\\_/ \\_____/")
	logger.Log("\n")
	logger.Log("\n")

}
