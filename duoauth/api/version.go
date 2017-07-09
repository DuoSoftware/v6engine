package api

import (
	"duov6.com/common"
	"encoding/json"
	"runtime"
	"strconv"
	"time"
)

//------------------------ Version Management --------------------------------

var StartTime time.Time

func Verify() (output string) {
	cpuUsage := strconv.Itoa(int(common.GetProcessorUsage()))
	cpuCount := strconv.Itoa(runtime.NumCPU())

	versionData := make(map[string]interface{})
	versionData["API Name"] = "Duo Auth ( Azure AD )"
	versionData["API Version"] = "1.0.0a"

	changeLogs := make(map[string]interface{})

	changeLogs["1.0.0"] = "Revamped begining of Azure AD DuoAuth."
	versionData["Change Logs"] = changeLogs

	statMap := make(map[string]string)
	statMap["CPU"] = cpuUsage + " (percentage)"
	statMap["CPU Cores"] = cpuCount
	nowTime := time.Now()
	elapsedTime := nowTime.Sub(StartTime)
	statMap["Time Started"] = StartTime.UTC().Add(330 * time.Minute).Format(time.RFC1123)
	statMap["Time Elapsed"] = elapsedTime.String()
	versionData["Metrics"] = statMap

	gitMap := make(map[string]string)
	gitMap["Type"] = "git"
	gitMap["URL"] = "https://github.com/DuoSoftware/v6engine/"
	versionData["Repository"] = gitMap

	authorMap := make(map[string]string)
	authorMap["Name"] = "Duo Software Pvt Ltd"
	authorMap["URL"] = "http://www.duosoftware.com/"
	versionData["Project Author"] = authorMap

	byteArray, _ := json.Marshal(versionData)
	output = string(byteArray)
	return
}
