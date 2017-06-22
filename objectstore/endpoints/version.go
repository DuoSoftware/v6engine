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

	versionData := make(map[string]interface{})
	versionData["API Name"] = "ObjectStore"
	versionData["API Version"] = "6.1.09"

	changeLogs := make(map[string]interface{})

	changeLogs["6.1.01"] = "Added timezone compatibility"
	changeLogs["6.1.02"] = "Added redis key update clear call"
	changeLogs["6.1.03"] = "Added MySQL JSON store."
	changeLogs["6.1.04"] = "Removed TimeZone compatability temporarily."
	changeLogs["6.1.05"] = "Added Toggle Logs and Removed Log header requirement."
	changeLogs["6.1.06"] = "Added test version for replying error messages."
	changeLogs["6.1.07"] = "Added more metrics."
	changeLogs["6.1.08"] = "Added wait for config before starting webservice."
	changeLogs["6.1.09"] = "Added Get GUID for special methods"
	versionData["ChangeLogs"] = changeLogs

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
