package endpoints

import (
	"duov6.com/common"
	"encoding/json"
	"runtime"
	"strconv"
	"time"
)

var StartTime time.Time

func GetVersion() string {
	cpuUsage := strconv.Itoa(int(common.GetProcessorUsage()))
	cpuCount := strconv.Itoa(runtime.NumCPU())
	//versionDaata := "{\"Name\": \"Objectstore\",\"Version\": \"1.4.4-a\",\"Change Log\":\"Fixed certain alter table issues.\",\"Author\": {\"Name\": \"Duo Software\",\"URL\": \"http://www.duosoftware.com/\"},\"Repository\": {\"Type\": \"git\",\"URL\": \"https://github.com/DuoSoftware/v6engine/\"},\"System Usage\": {\"CPU\": \" " + cpuUsage + " (percentage)\",\"CPU Cores\": \"" + cpuCount + "\"}}"
	versionData := make(map[string]interface{})
	versionData["API Name"] = "Duo Notifier"
	versionData["API Version"] = "6.1.01"

	changeLogs := make(map[string]interface{})
	changeLogs["6.1.01"] = "Added metrics and Environment variable Settings file generation."
	changeLogs["6.1.00"] = "Started new versioning with 6.1.00, Added agent.config to reflect localhost if agent.config not found"
	versionData["changeLogs"] = changeLogs

	statMap := make(map[string]string)
	statMap["CPU"] = cpuUsage + " (percentage)"
	statMap["CPU Cores"] = cpuCount
	nowTime := time.Now()
	elapsedTime := nowTime.Sub(StartTime)
	statMap["Time Started"] = StartTime.UTC().Add(330 * time.Minute).Format(time.RFC1123)
	statMap["Time Elapsed"] = elapsedTime.String()
	versionData["Metrics"] = statMap

	authorMap := make(map[string]string)
	authorMap["Name"] = "Duo Software Pvt Ltd"
	authorMap["URL"] = "http://www.duosoftware.com/"
	versionData["Project Author"] = authorMap

	gitMap := make(map[string]string)
	gitMap["Type"] = "git"
	gitMap["URL"] = "https://github.com/DuoSoftware/v6engine/"
	versionData["Repository"] = gitMap

	byteArray, _ := json.Marshal(versionData)
	return string(byteArray)
}
