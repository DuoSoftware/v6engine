package api

import (
	"duov6.com/cebadapter"
	"duov6.com/common"
	"duov6.com/duoauth/azureapi"
	// notifier "duov6.com/duonotifier/client"
	// "duov6.com/objectstore/client"
	// "duov6.com/session"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"github.com/SiyaDlamini/gorest"
	"net/url"
	// "strconv"
	"errors"
	"strings"
)

type Auth struct {
	gorest.RestService
	IsServiceReferral bool            // if the referal is a service based one dont check for session
	verify            gorest.EndPoint `method:"GET" path:"/" output:"string"`
	getConfig         gorest.EndPoint `method:"GET" path:"/config" output:"string"`
	getSession        gorest.EndPoint `method:"GET" path:"/getsession" output:"AuthResponse"`
	getUser           gorest.EndPoint `method:"GET" path:"/users/{Email:string}" output:"AuthResponse"`
	createUser        gorest.EndPoint `method:"POST" path:"/users" postdata:"UserCreateInfo"`
	updateUser        gorest.EndPoint `method:"PUT" path:"/users" postdata:"UserCreateInfo"`
	deleteUser        gorest.EndPoint `method:"DELETE" path:"/users/{Email:string}"`
	noIdpProcess      gorest.EndPoint `method:"GET" path:"/users/noidp" output:"AuthResponse"`
	//scope management
	assignUserScopes gorest.EndPoint `method:"POST" path:"/users/scopes/{Email:string}" postdata:"[]string"`
	//logs
	toggleLogs gorest.EndPoint `method:"GET" path:"/togglelogs/" output:"string"`
}

var agentConfig map[string]interface{}

func (A Auth) GetSession() AuthResponse {
	term.Write("Executing Method : Get Session ", term.Blank)
	response := AuthResponse{}

	var err error

	if agentConfig == nil {
		agentConfig = make(map[string]interface{})
		agentConfig = common.VerifyConfigFiles()
	}

	id_token := A.Context.Request().Header.Get("Securitytoken")
	if id_token != "" {
		urlFragment := "azure.smoothflow.io"
		//urlFragment := agentConfig["objUrl"].(string)
		urlFragment = strings.Replace(urlFragment, ":3000", "", -1)
		urlFragment = strings.Replace(urlFragment, "https://", "", -1)
		urlFragment = strings.Replace(urlFragment, "http://", "", -1)
		graphUrl := "https://" + urlFragment + "/auth/GetSession"
		fmt.Println(graphUrl)

		headers := make(map[string]string)
		headers["Securitytoken"] = id_token
		headers["Content-Type"] = "application/json"

		var body []byte
		err, body = common.HTTP_GET(graphUrl, headers, false)
		if err == nil {
			_ = json.Unmarshal(body, &response)
			response.Status = true
			response.Message = "Session recieved successfully."
			response.Data = response
		} else {
			fmt.Println(string(body))
			var newResponse AuthResponse
			_ = json.Unmarshal(body, &newResponse)
			response.Status = false
			response.Message = newResponse.Message
		}
	} else {
		response.Status = false
		response.Message = "SecurityToken not found in header."
	}

	return response
}

func (A Auth) NoIdpProcess() AuthResponse {
	term.Write("Executing Method : Process No IDP user", term.Blank)
	response := AuthResponse{}

	var err error
	id_token := A.Context.Request().Header.Get("Securitytoken")
	if id_token != "" {
		//get session
		sesResp := A.GetSession()
		if sesResp.Status {
			//correct request.. fetch profile from AAD
			access_token, err := azureapi.GetGraphApiToken()
			if err == nil {
				objectID := sesResp.Data.(AuthResponse).Data.(map[string]interface{})["oid"].(string)
				email := sesResp.Data.(AuthResponse).Data.(map[string]interface{})["emails"].([]interface{})[0].(string)

				graphUrl := "https://graph.windows.net/smoothflowio.onmicrosoft.com/users/" + objectID + "?api-version=1.6"
				headers := make(map[string]string)
				headers["Authorization"] = "Bearer " + access_token
				headers["Content-Type"] = "application/json"

				//update email
				jsonData := `"otherMails": ["` + email + `"]`

				err, _ = common.HTTP_PATCH(graphUrl, headers, []byte(jsonData), false)
				if err == nil {
					response.Status = true
					response.Message = "Successfullt processed no idp user."
				}
			}
		} else {
			err = errors.New(sesResp.Message)
		}

	} else {
		err = errors.New("Securitytoken not found in header.")
	}

	if err != nil {
		response.Status = false
		response.Message = err.Error()
	}

	return response
}

func (A Auth) GetUser(Email string) AuthResponse {
	term.Write("Executing Method : Get User", term.Blank)
	response := AuthResponse{}

	var err error
	id_token := A.Context.Request().Header.Get("Securitytoken")
	if id_token != "" {
		//correct request.. fetch profile from AAD
		access_token, err := azureapi.GetGraphApiToken()
		if err == nil {
			graphUrl := "https://graph.windows.net/smoothflowio.onmicrosoft.com/users?api-version=1.6&$filter=otherMails/any" + url.QueryEscape("(o: o eq '"+Email+"')")
			headers := make(map[string]string)
			headers["Authorization"] = "Bearer " + access_token
			headers["Content-Type"] = "application/json"

			var body []byte
			err, body = common.HTTP_GET(graphUrl, headers, false)
			if err == nil {
				data := make(map[string]interface{})
				_ = json.Unmarshal(body, &data)
				data = data["value"].([]interface{})[0].(map[string]interface{})
				user := User{}
				user.EmailAddress = Email
				user.Name = data["displayName"].(string)
				user.Country = data["country"].(string)
				user.ObjectID = data["objectId"].(string)

				if data["jobTitle"] != nil {
					user.Scopes = strings.Split(data["jobTitle"].(string), "-")
				}

				tenantString := ""
				if data["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant"] != nil {
					tenantString += data["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant"].(string)
				}
				if data["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant1"] != nil {
					tenantString += "-" + data["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant1"].(string)
				}
				if data["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant2"] != nil {
					tenantString += "-" + data["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant2"].(string)
				}
				if data["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant3"] != nil {
					tenantString += "-" + data["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant3"].(string)
				}
				if data["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant4"] != nil {
					tenantString += "-" + data["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant4"].(string)
				}

				alltenants := strings.Split(tenantString, "-")
				userTenant := make([]UserTenant, len(alltenants))
				for x := 0; x < len(alltenants); x++ {
					entry := alltenants[x]
					singleTenant := UserTenant{}
					if strings.Contains(entry, "default#") {
						singleTenant.IsDefault = true
						entry = strings.Replace(entry, "default#", "", -1)
					}
					if strings.Contains(entry, "admin#") {
						singleTenant.IsAdmin = true
						entry = strings.Replace(entry, "admin#", "", -1)
					}
					singleTenant.TenantID = entry
					userTenant[x] = singleTenant
				}
				user.Tenants = userTenant

				response.Status = true
				response.Message = "User profile recieved successfully."
				response.Data = user
			}
		}

	} else {
		response.Status = false
		response.Message = "Securitytoken not found in header."
	}

	if err != nil {
		response.Status = false
		response.Message = err.Error()
	}

	return response
}

func (A Auth) CreateUser(u UserCreateInfo) {
	term.Write("Executing Method : Create a local user.", term.Blank)
	response := AuthResponse{}
	access_token, err := azureapi.GetGraphApiToken()
	if err == nil {
		//create local user
		graphUrl := "https://graph.windows.net/smoothflowio.onmicrosoft.com/users?api-version=1.6"
		headers := make(map[string]string)
		headers["Authorization"] = "Bearer " + access_token
		headers["Content-Type"] = "application/json"

		if u.Password == "" {
			u.Password = common.GetGUID()
		}

		jsonString := `{
  "accountEnabled": true,
  "creationType": "LocalAccount",
  "displayName": "` + u.Name + `",
  "country": "` + u.Country + `",
  "otherMails": ["` + u.Email + `"],
  "passwordProfile": {
    "password": "` + u.Password + `",
    "forceChangePasswordNextLogin": false
  },
  "signInNames": [
    {
      "type": "userName",
      "value": "` + u.Name + `"
    },
    {
      "type": "emailAddress",
      "value": "` + u.Email + `"
    }
  ]
}`

		err, _ = common.HTTP_POST(graphUrl, headers, []byte(jsonString), false)

		if err == nil && u.TenantID != "" { //if user creation success and tenantid is not nil
			//assign user scopes
			A.IsServiceReferral = true
			scopes := strings.Split("B-FO-FS-DD", "-")
			A.AssignUserScopes(scopes, u.Email)
			A.IsServiceReferral = false
			//assign to tenant if available.
			//get tenant objectID
			T := TenantSvc{}
			T.RestService.Context = A.Context
			T.IsServiceReferral = true
			tenantResp := T.GetTenant(u.TenantID)
			if tenantResp.Status {
				T.RestService.Context.Request().Header.Set("Nounce", u.TenantID)
				addUserResp := T.AddUserToTenant(u.TenantID, u.Email)
				if addUserResp.Status {
					response.Status = true
					response.Message = "User created successfully and added to tenant."
				} else {
					err = errors.New(addUserResp.Message)
				}
			} else {
				if tenantResp.Message == "Tenant not found." {
					//this is first time. //create the tenant
					tenant := Tenant{}
					tenant.Admin = u.Email
					tenant.Country = u.Country
					tenant.TenantID = u.TenantID
					tenant.Type = "JIRA"
					T.CreateTenant(tenant)
					//assign user to tenant.
					T.RestService.Context.Request().Header.Set("Nounce", "defaultNonce")
					addUserResp := T.AddUserToTenant(u.TenantID, u.Email)
					if addUserResp.Status {
						response.Status = true
						response.Message = "User and tenant created successfully and added to tenant."
					} else {
						err = errors.New(addUserResp.Message)
					}
				} else {
					err = errors.New(tenantResp.Message)
				}
			}
		} else {
			response.Status = true
			response.Message = "User created but since no tenant details were supplied, not assigned to a tenant. Run /tenants/{tid:string}/adduser/{Email:string} to add this user to a tenant."
		}
	}

	if err != nil {
		fmt.Println(err.Error())
		response.Status = false
		response.Message = err.Error()
		b, _ := json.Marshal(response)
		A.ResponseBuilder().SetResponseCode(500).WriteAndOveride(b)
	} else {
		b, _ := json.Marshal(response)
		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
	}
}

func (A Auth) UpdateUser(u UserCreateInfo) {
	term.Write("Executing Method : Update local user.", term.Blank)
	response := AuthResponse{}

	var err error
	id_token := A.Context.Request().Header.Get("Securitytoken")

	if id_token != "" {
		sessionResponse := A.GetSession()
		if sessionResponse.Status {
			data := sessionResponse.Data.(AuthResponse).Data.(map[string]interface{})
			userObjectID := data["oid"].(string)
			name := ""
			country := ""
			if u.Name != "" {
				name = u.Name
			}
			if u.Country != "" {
				country = u.Country
			}

			if (name == "" && country == "") || (name == data["name"].(string) && name == data["country"].(string)) {
				err = errors.New("No new information to be updated.")
			} else {
				var access_token string
				access_token, err = azureapi.GetGraphApiToken()
				if err == nil {
					graphUrl := "https://graph.windows.net/smoothflowio.onmicrosoft.com/users/" + userObjectID + "?api-version=1.6"
					headers := make(map[string]string)
					headers["Authorization"] = "Bearer " + access_token
					headers["Content-Type"] = "application/json"

					jsonString := `{"displayName":"` + name + `", "country":"` + country + `"}`
					err, _ = common.HTTP_PATCH(graphUrl, headers, []byte(jsonString), false)
				}
			}

		} else {
			err = errors.New(sessionResponse.Message)
		}
	} else {
		err = errors.New("Securitytoken not found in header.")
	}

	if err != nil {
		fmt.Println(err.Error())
		response.Status = false
		response.Message = err.Error()
		b, _ := json.Marshal(response)
		A.ResponseBuilder().SetResponseCode(500).WriteAndOveride(b)
	} else {
		b, _ := json.Marshal(response)
		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
	}
}

func (A Auth) DeleteUser(Email string) {
	term.Write("Executing Method : Delete user.", term.Blank)
	response := AuthResponse{}
	response.Status = false
	response.Message = "Not supported yet."
	b, _ := json.Marshal(response)
	A.ResponseBuilder().SetResponseCode(501).WriteAndOveride(b)
}

func (A Auth) AssignUserScopes(scopes []string, Email string) {
	term.Write("Executing Method : Assign User Scope", term.Blank)
	response := AuthResponse{}

	scopeMap := make(map[string]interface{})
	for x := 0; x < len(scopes); x++ {
		scopeMap[scopes[x]] = "ignoreValue"
	}

	var err error
	id_token := A.Context.Request().Header.Get("Securitytoken")
	if A.IsServiceReferral || id_token != "" {
		access_token, err := azureapi.GetGraphApiToken()
		if err == nil {
			//fetch user
			getUserResponse := A.GetUser(Email)
			if !getUserResponse.Status {
				err = errors.New(getUserResponse.Message)
			} else {
				currentScopes := (getUserResponse.Data).(User).Scopes
				if len(currentScopes) != 0 {
					for x := 0; x < len(currentScopes); x++ {
						if scopeMap[currentScopes[x]] == nil {
							scopeMap[currentScopes[x]] = "ignoreValue"
						}
					}
				}

				scopeString := ""
				for key, _ := range scopeMap {
					scopeString += "-" + key
				}

				fmt.Println(scopeString)

				scopeString = strings.TrimPrefix(scopeString, "-")

				//update the user
				graphUrl := "https://graph.windows.net/smoothflowio.onmicrosoft.com/users/" + (getUserResponse.Data).(User).ObjectID + "?api-version=1.6"

				headers := make(map[string]string)
				headers["Authorization"] = "Bearer " + access_token
				headers["Content-Type"] = "application/json"

				jsonString := `{"jobTitle": "` + scopeString + `"}`

				err, _ = common.HTTP_PATCH(graphUrl, headers, []byte(jsonString), false)
				if err == nil {
					response.Status = true
					response.Message = "Profile scopes assigned successfully."
				}
			}
		}
	} else {
		err = errors.New("No Securitytoken found in header.")
	}

	if err != nil {
		fmt.Println(err.Error())
		response.Status = false
		response.Message = err.Error()
		b, _ := json.Marshal(response)
		A.ResponseBuilder().SetResponseCode(500).WriteAndOveride(b)
	} else {
		b, _ := json.Marshal(response)
		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
	}
}

//.......................................

func (A Auth) ToggleLogs() string {
	return term.ToggleConfig()
}

func (A Auth) GetConfig() (output string) {
	configAll := cebadapter.GetGlobalConfig("StoreConfig")
	byteArray, _ := json.Marshal(configAll)
	return string(byteArray)
}

func (A Auth) Verify() (output string) {
	output = Verify()
	return
}
