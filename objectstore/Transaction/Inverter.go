package Transaction

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"duov6.com/objectstore/storageengines"
	"encoding/json"
	"strconv"
	"strings"
)

func GetInvertedRequests(request *messaging.ObjectRequest) (retRequests []*messaging.ObjectRequest) {
	//if delete.. get records for keys and store as a Insert
	//if insert.. check if its there. get that record and store as an insert bcs its going to be updated.
	//			  if not.. store as a delete.
	//if update.. same as above.

	originalOperation := strings.ToLower(request.Controls.Operation)

	switch originalOperation {
	case "insert":
		break
	case "update":
		break
	case "delete":
		break
	default:
		//no change.. Append nothing!
		break

	}

	return
}

func GetDeleteInverted(request *messaging.ObjectRequest) (retRequest *messaging.ObjectRequest) {
	deleteRequest := messaging.ObjectRequest{}
	deleteRequest.Controls = request.Controls
	//deleteRequest.Body = request.Body
	deleteRequest.Configuration = request.Configuration
	deleteRequest.Extras = request.Extras
	deleteRequest.IsLogEnabled = request.IsLogEnabled
	deleteRequest.MessageStack = request.MessageStack
	deleteRequest.Controls.Operation = "insert"

	if request.Body.Object != nil {
		getRequest := messaging.ObjectRequest{}
		getRequest.Controls = request.Controls
		getRequest.Configuration = request.Configuration
		getRequest.Extras = request.Extras
		getRequest.IsLogEnabled = false
		getRequest.Controls.Operation = "read-key"
		getRequest.Controls.Id = request.Body.Object[request.Body.Parameters.KeyProperty].(string)

		response := ProcessRequest(&getRequest)

		if response.IsSuccess {
			var Object map[string]interface{}
			_ = json.Unmarshal(response.Body, &Object)
			deleteRequest.Body.Object = Object
		}
	} else {
		var allDeletes []map[string]interface{}
		for _, singleDelete := range request.Body.Objects {
			getRequest := messaging.ObjectRequest{}
			getRequest.Controls = request.Controls
			getRequest.Configuration = request.Configuration
			getRequest.Extras = request.Extras
			getRequest.IsLogEnabled = false
			getRequest.Controls.Operation = "read-key"
			getRequest.Controls.Id = singleDelete[request.Body.Parameters.KeyProperty].(string)

			response := ProcessRequest(&getRequest)

			if response.IsSuccess {
				var Object map[string]interface{}
				_ = json.Unmarshal(response.Body, &Object)
				allDeletes = append(allDeletes, Object)
			}
		}

		if allDeletes != nil {
			deleteRequest.Body.Objects = allDeletes
		}
	}

	return
}

func ProcessRequest(request *messaging.ObjectRequest) repositories.RepositoryResponse {

	var storageEngine storageengines.AbstractStorageEngine // request.StoreConfiguration.StorageEngine
	storageEngine = storageengines.ReplicatedStorageEngine{}

	var outResponse repositories.RepositoryResponse = storageEngine.Store(request)

	if request.IsLogEnabled {
		for index, element := range request.MessageStack {
			request.Log("S-" + strconv.Itoa(index) + " : " + element)
		}
	}

	return outResponse
}
