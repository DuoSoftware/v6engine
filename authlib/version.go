package authlib

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
	//output = "{\"name\": \"DuoAuth\",\"version\": \"6.0.24-a\",\"Change Log\":\"Added Check for tenant subscription invitation.\",\"author\": {\"name\": \"Duo Software\",\"url\": \"http://www.duosoftware.com/\"},\"repository\": {\"type\": \"git\",\"url\": \"https://github.com/DuoSoftware/v6engine/\"}}"
	cpuUsage := strconv.Itoa(int(common.GetProcessorUsage()))
	cpuCount := strconv.Itoa(runtime.NumCPU())

	versionData := make(map[string]interface{})
	versionData["API Name"] = "Duo Auth"
	versionData["API Version"] = "6.2.02c"

	changeLogs := make(map[string]interface{})

	changeLogs["6.2.02"] = "Added REST CEB configuration retrieval."
	changeLogs["6.2.01"] = "Added wait for config before starting webservice."
	changeLogs["6.2.00"] = "Added check login method"
	changeLogs["6.1.30"] = "Added settings.config generator from ENV values"
	changeLogs["6.1.29"] = "Added Cancel tenant invitation."
	changeLogs["6.1.28"] = "Added account deactive, reactive and delete functions."
	changeLogs["6.1.27"] = "Fixed a bug in multiple sessions."
	changeLogs["6.1.26"] = "Added more metrics."
	changeLogs["6.1.25"] = "Added response changes."
	changeLogs["6.1.24"] = "Added checks for non session timeout instances."
	changeLogs["6.1.23"] = "Automatic Session Timeouts."
	changeLogs["6.1.22"] = "Added invite based tenant regs"
	changeLogs["6.1.21"] = "Added Resend activation email."
	changeLogs["6.1.20"] = "Removed securityToken check being applied twice at GetTenants."
	changeLogs["6.1.19"] = "Sorted GetTenant for UserID method to output default tenant in the first index."
	changeLogs["6.1.18"] = "Fixed a security hole. Now only admins can remove users."
	changeLogs["6.1.17"] = "Added URL based Password Reset"
	changeLogs["6.1.16"] = "Added User Activation By Tenant Admin and removed auto activation for Custom Tenant User Registration."
	changeLogs["6.1.15"] = "Added Reset password by tenant admin."
	changeLogs["6.1.14"] = "Removed SecurityToken check for GetTenant(), Added sessionless User registration with TenantID."
	changeLogs["6.1.13"] = "Added GetAllPendingTenantRequests Method."
	changeLogs["6.1.12"] = "Changed GetTenantAdmin to get All Admin data. Fixed tenant invite for existing customer issue."
	changeLogs["6.1.11"] = "Fixed few email template issues. JIRA : EX-1085"
	changeLogs["6.1.10"] = "Added Toggle Logs and disabled CMD logs at startup. User /ToggleLogs to cycle through different logs."
	changeLogs["6.1.09"] = "Added new user email templates for events."
	changeLogs["6.1.08"] = "Added user deny check, Added User Deactivate if user has no accesible tenants."
	changeLogs["6.1.07"] = "Added Activation Skip Endpoint for Registration. <InvitedUserRegistration>"
	changeLogs["6.1.06"] = "Commented SecurityToken from AcceptRequest, Added response codes for ActivateUser method"
	changeLogs["6.1.05"] = "Added New Login password,username message and Activate message, Added GetTenantAdmin method for auth, Removed rating engine check for tenant add."
	changeLogs["6.1.04"] = "Added Activate User Email Check. , Added Reset Password Check by checking user activated or not"
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
