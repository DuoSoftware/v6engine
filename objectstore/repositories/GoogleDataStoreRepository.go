package repositories

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/queryparser"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"github.com/twinj/uuid"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/datastore"
	"reflect"
	"strconv"
	"strings"
)

type GoogleDataStoreRepository struct {
}

func (repository GoogleDataStoreRepository) GetRepositoryName() string {
	return "GoogleDataStore"
}

func (repository GoogleDataStoreRepository) getConnection(request *messaging.ObjectRequest) (client *datastore.Client, err error) {
	dataStoreConfig := request.Configuration.ServerConfiguration["GoogleDataStore"]
	projectID := dataStoreConfig["ProjectID"]

	var key map[string]string
	key = make(map[string]string)
	key["type"] = dataStoreConfig["type"]
	key["private_key_id"] = dataStoreConfig["private_key_id"]
	key["private_key"] = dataStoreConfig["private_key"]
	key["client_email"] = dataStoreConfig["client_email"]
	key["client_id"] = dataStoreConfig["client_id"]
	key["auth_uri"] = dataStoreConfig["auth_uri"]
	key["token_uri"] = dataStoreConfig["token_uri"]
	key["auth_provider_x509_cert_url"] = dataStoreConfig["auth_provider_x509_cert_url"]
	key["client_x509_cert_url"] = dataStoreConfig["client_x509_cert_url"]

	jsonKey := getByteByValue(key)

	conf, err := google.JWTConfigFromJSON(
		jsonKey,
		datastore.ScopeDatastore,
		datastore.ScopeUserEmail,
	)
	if err != nil {
		term.Write(err.Error(), 1)
	} else {
		ctx := context.Background()
		client, err = datastore.NewClient(ctx, projectID, cloud.WithTokenSource(conf.TokenSource(ctx)))
		if err != nil {
			term.Write(err.Error(), 1)
		}
	}
	return
}

func (repository GoogleDataStoreRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-ALL")
	response := RepositoryResponse{}

	isOrderByAsc := false
	isOrderByDesc := false
	orderbyfield := ""

	skip := 0
	take := 100

	if request.Extras["skip"] != nil {
		if intValue, err := strconv.Atoi(request.Extras["skip"].(string)); err == nil {
			skip = intValue
		}
	}
	if request.Extras["take"] != nil {
		if intValue, err := strconv.Atoi(request.Extras["take"].(string)); err == nil {
			take = intValue
		}
	}
	if request.Extras["orderby"] != nil {
		orderbyfield = request.Extras["orderby"].(string)
		isOrderByAsc = true
	} else if request.Extras["orderbydsc"] != nil {
		orderbyfield = request.Extras["orderbydsc"].(string)
		isOrderByDesc = true
	}

	ctx := context.Background()
	client, err := repository.getConnection(request)
	ctx = datastore.WithNamespace(ctx, request.Controls.Namespace)

	if err != nil {
		fmt.Println(err.Error())
	} else {

		props := make([]datastore.PropertyList, 0)
		var data []map[string]interface{}

		var query *datastore.Query

		if isOrderByAsc {
			query = datastore.NewQuery(request.Controls.Class).Offset(skip).Limit(take).Order(orderbyfield)
		} else if isOrderByDesc {
			query = datastore.NewQuery(request.Controls.Class).Offset(skip).Limit(take).Order(("-" + orderbyfield))
		} else {
			query = datastore.NewQuery(request.Controls.Class).Offset(skip).Limit(take)
		}

		_, err := client.GetAll(ctx, query, &props)
		if err != nil {
			term.Write(err.Error(), 1)
			response.GetResponseWithBody(getEmptyByteObject())
		} else {
			//data recieved! :)
			for index := 0; index < len(props); index++ {
				var record map[string]interface{}
				record = make(map[string]interface{})
				for _, value := range props[index] {
					if value.Name != "_os_id" && value.Name != "__osHeaders" {
						record[value.Name] = repository.GQLToGolang(value.Value)
					}
				}
				data = append(data, record)
			}
		}

		bytesValue, _ := json.Marshal(data)
		if len(bytesValue) == 4 || len(bytesValue) == 2 {
			bytesValue = getEmptyByteObject()
		}

		response.IsSuccess = true
		response.Message = "Values Retrieved Successfully from Google DataStore!"
		response.GetResponseWithBody(bytesValue)
	}
	return response
}

func (repository GoogleDataStoreRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-Search!", 2)
	response := RepositoryResponse{}

	isOrderByAsc := false
	isOrderByDesc := false
	orderbyfield := ""

	skip := 0
	take := 100

	if request.Extras["skip"] != nil {
		if intValue, err := strconv.Atoi(request.Extras["skip"].(string)); err == nil {
			skip = intValue
		}
	}
	if request.Extras["take"] != nil {
		if intValue, err := strconv.Atoi(request.Extras["take"].(string)); err == nil {
			take = intValue
		}
	}
	if request.Extras["orderby"] != nil {
		orderbyfield = request.Extras["orderby"].(string)
		isOrderByAsc = true
	} else if request.Extras["orderbydsc"] != nil {
		orderbyfield = request.Extras["orderbydsc"].(string)
		isOrderByDesc = true
	}

	ctx := context.Background()
	client, err := repository.getConnection(request)
	ctx = datastore.WithNamespace(ctx, request.Controls.Namespace)

	var query *datastore.Query
	if strings.Contains(request.Body.Query.Parameters, ":") {
		tokens := strings.Split(request.Body.Query.Parameters, ":")
		fieldName := tokens[0]
		fieldValue := tokens[1]
		fieldName = strings.TrimSpace(fieldName)
		fieldValue = strings.TrimSpace(fieldValue)
		fmt.Println(fieldValue)
		fmt.Println(fieldName)
		if isOrderByAsc {
			query = datastore.NewQuery(request.Controls.Class).Filter((fieldName + " ="), repository.getSearchToken(fieldValue)).Offset(skip).Limit(take).Order(orderbyfield)
		} else if isOrderByDesc {
			query = datastore.NewQuery(request.Controls.Class).Filter((fieldName + " ="), repository.getSearchToken(fieldValue)).Offset(skip).Limit(take).Order(("-" + orderbyfield))
		} else {
			query = datastore.NewQuery(request.Controls.Class).Filter((fieldName + " ="), repository.getSearchToken(fieldValue)).Offset(skip).Limit(take)
		}
	} else {
		query = datastore.NewQuery(request.Controls.Class).Offset(skip).Limit(take)
	}

	if err != nil {
		fmt.Println(err.Error())
	} else {

		props := make([]datastore.PropertyList, 0)
		var data []map[string]interface{}

		_, err := client.GetAll(ctx, query, &props)
		if err != nil {
			response.GetResponseWithBody(getEmptyByteObject())
			term.Write(err.Error(), 1)
		} else {
			//data recieved! :)
			for index := 0; index < len(props); index++ {
				var record map[string]interface{}
				record = make(map[string]interface{})
				for _, value := range props[index] {
					if value.Name != "_os_id" && value.Name != "__osHeaders" {
						record[value.Name] = repository.GQLToGolang(value.Value)
					}
				}
				data = append(data, record)
			}
		}

		bytesValue, _ := json.Marshal(data)
		if len(bytesValue) == 4 || len(bytesValue) == 2 {
			bytesValue = getEmptyByteObject()
		}

		response.IsSuccess = true
		response.Message = "Values Retrieved Successfully from Google DataStore!"
		response.GetResponseWithBody(bytesValue)
	}
	return response
}

func (repository GoogleDataStoreRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-QUERY!")
	response := RepositoryResponse{}
	queryType := request.Body.Query.Type

	switch queryType {
	case "Query":
		if request.Body.Query.Parameters != "*" {
			ctx := context.Background()
			client, err := repository.getConnection(request)
			ctx = datastore.WithNamespace(ctx, request.Controls.Namespace)

			if err != nil {
				response.Message = "Values Retrieved Successfully from Google DataStore!"
				response.GetResponseWithBody(getEmptyByteObject())
				return response
			} else {
				props := make([]datastore.PropertyList, 0)
				var data []map[string]interface{}

				var query *datastore.Query

				query, qErr := queryparser.GetDataStoreQuery(request.Body.Query.Parameters, request.Controls.Namespace, request.Controls.Class)
				if qErr != nil {
					response.Message = "Values Retrieved Successfully from Google DataStore!"
					response.GetResponseWithBody(getEmptyByteObject())
					return response
				}

				fmt.Print("Normalized Data Store Query : ")
				fmt.Println(query)

				_, err := client.GetAll(ctx, query, &props)
				if err != nil {
					term.Write(err.Error(), 1)
					response.Message = "Values Retrieved Successfully from Google DataStore!"
					response.GetResponseWithBody(getEmptyByteObject())
					return response
				} else {
					//data recieved! :)
					for index := 0; index < len(props); index++ {
						var record map[string]interface{}
						record = make(map[string]interface{})
						for _, value := range props[index] {
							if value.Name != "_os_id" && value.Name != "__osHeaders" {
								record[value.Name] = repository.GQLToGolang(value.Value)
							}
						}
						data = append(data, record)
					}
				}

				bytesValue, _ := json.Marshal(data)
				if len(bytesValue) == 4 || len(bytesValue) == 2 {
					bytesValue = getEmptyByteObject()
				}

				response.IsSuccess = true
				response.Message = "Values Retrieved Successfully from Google DataStore!"
				response.GetResponseWithBody(bytesValue)
			}
		} else {
			return repository.GetAll(request)
		}
	default:
		request.Log(queryType + " is not implemented in Google DataStore Db repository")
		return getDefaultNotImplemented()
	}
	return response
}

func (repository GoogleDataStoreRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-BY-KEY")
	response := RepositoryResponse{}

	ctx := context.Background()
	client, err := repository.getConnection(request)
	ctx = datastore.WithNamespace(ctx, request.Controls.Namespace)

	if err != nil {
		fmt.Println(err.Error())
		response.GetResponseWithBody(getEmptyByteObject())
	} else {
		key := datastore.NewKey(ctx, request.Controls.Class, getNoSqlKey(request), 0, nil)

		var props datastore.PropertyList
		var data map[string]interface{}
		data = make(map[string]interface{})

		if err := client.Get(ctx, key, &props); err != nil {
			term.Write(err.Error(), 1)
		} else {
			for _, value := range props {
				if value.Name != "_os_id" && value.Name != "__osHeaders" {
					data[value.Name] = repository.GQLToGolang(value.Value)
				}
			}
		}

		bytesValue, _ := json.Marshal(data)
		if len(bytesValue) == 4 || len(bytesValue) == 2 {
			bytesValue = getEmptyByteObject()
		}

		response.IsSuccess = true
		response.Message = "Values Retrieved Successfully from Google DataStore!"
		response.GetResponseWithBody(bytesValue)

	}
	return response
}

func (repository GoogleDataStoreRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-MULTIPLE")
	return repository.setManyDataStore(request)
}

func (repository GoogleDataStoreRepository) setManyDataStore(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	var idData map[string]interface{}
	idData = make(map[string]interface{})

	ctx := context.Background()
	client, err := repository.getConnection(request)
	ctx = datastore.WithNamespace(ctx, request.Controls.Namespace)

	if err == nil {

		for index, obj := range request.Body.Objects {
			id := repository.getRecordID(request, obj)
			idData[strconv.Itoa(index)] = id
			request.Body.Objects[index][request.Body.Parameters.KeyProperty] = id
		}

		var keys []*datastore.Key
		keys = make([]*datastore.Key, len(request.Body.Objects))

		propArray := make([]interface{}, len(request.Body.Objects))

		for index := 0; index < len(request.Body.Objects); index++ {
			keys[index] = datastore.NewKey(ctx, request.Controls.Class, getNoSqlKeyById(request, request.Body.Objects[index]), 0, nil)
			var props datastore.PropertyList
			props = datastore.PropertyList{}
			props = append(props, datastore.Property{Name: "_os_id", Value: getNoSqlKeyById(request, request.Body.Objects[index])})
			for key, value := range request.Body.Objects[index] {
				props = append(props, datastore.Property{Name: key, Value: repository.GolangToGQL(value)})
			}
			propArray[index] = &props
		}

		if _, err := client.PutMulti(ctx, keys, propArray); err != nil {
			request.Log(err.Error())
			response.IsSuccess = false
			response.Message = "Error storing object in Google DataStore : " + err.Error()
		} else {
			response.IsSuccess = true
			response.Message = "Successfully stored object in Google DataStore"
		}

	} else {
		request.Log("Connection Failed!")
		response.IsSuccess = false
		response.Message = "Connection Failed to Google DataStore"
	}

	var DataMap []map[string]interface{}
	DataMap = make([]map[string]interface{}, 1)
	var idMap map[string]interface{}
	idMap = make(map[string]interface{})
	idMap["ID"] = idData
	DataMap[0] = idMap
	response.Data = DataMap

	return response
}

func (repository GoogleDataStoreRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-SINGLE")
	return repository.setOneDataStore(request)
}

func (repository GoogleDataStoreRepository) setOneDataStore(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	id := repository.getRecordID(request, request.Body.Object)
	request.Controls.Id = id
	request.Body.Object[request.Body.Parameters.KeyProperty] = id

	ctx := context.Background()
	client, err := repository.getConnection(request)
	ctx = datastore.WithNamespace(ctx, request.Controls.Namespace)

	key := datastore.NewKey(ctx, request.Controls.Class, getNoSqlKey(request), 0, nil)

	var props datastore.PropertyList
	props = append(props, datastore.Property{Name: "_os_id", Value: getNoSqlKey(request)})
	for key, value := range request.Body.Object {
		props = append(props, datastore.Property{Name: key, Value: repository.GolangToGQL(value)})
	}

	_, err = client.Put(ctx, key, &props)
	if err != nil {
		response.IsSuccess = false
		response.Message = "Error Insert/Update Object in Google DataStore! : " + err.Error()
		request.Log(err.Error())
	} else {
		response.IsSuccess = true
		response.Message = "Successfully stored object in Google DataStore"
	}

	//Add IDs to return Data
	var Data []map[string]interface{}
	Data = make([]map[string]interface{}, 1)
	var idData map[string]interface{}
	idData = make(map[string]interface{})
	idData["ID"] = id
	Data[0] = idData
	response.Data = Data
	return response
}

func (repository GoogleDataStoreRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-MULTIPLE")
	return repository.setManyDataStore(request)
}

func (repository GoogleDataStoreRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-SINGLE")
	return repository.setOneDataStore(request)
}

func (repository GoogleDataStoreRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-MULTIPLE")
	response := RepositoryResponse{}

	ctx := context.Background()
	client, err := repository.getConnection(request)
	if err == nil {
		ctx = datastore.WithNamespace(ctx, request.Controls.Namespace)

		var keys []*datastore.Key
		keys = make([]*datastore.Key, len(request.Body.Objects))

		for index, obj := range request.Body.Objects {
			keys[index] = datastore.NewKey(ctx, request.Controls.Class, getNoSqlKeyById(request, obj), 0, nil)
		}

		if err := client.DeleteMulti(ctx, keys); err != nil {
			response.IsSuccess = false
			response.Message = "Error deleting objects in Google DataStore : " + err.Error()
			request.Log(err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Success deleting objects in Google DataStore!"
		}
	} else {
		response.IsSuccess = false
		response.Message = "No Connection to DataStore : " + err.Error()
		request.Log(err.Error())
	}

	return response
}

func (repository GoogleDataStoreRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-SINGLE")
	response := RepositoryResponse{}

	ctx := context.Background()
	client, err := repository.getConnection(request)
	if err == nil {
		ctx = datastore.WithNamespace(ctx, request.Controls.Namespace)

		key := datastore.NewKey(ctx, request.Controls.Class, getNoSqlKey(request), 0, nil)

		if err := client.Delete(ctx, key); err != nil {
			response.IsSuccess = false
			response.Message = "Error deleting object in Google DataStore : " + err.Error()
			request.Log(err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Success deleting object in Google DataStore!"
		}
	} else {
		response.IsSuccess = false
		response.Message = "No Connection to DataStore : " + err.Error()
		request.Log(err.Error())
	}
	return response
}

func (repository GoogleDataStoreRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Special!", 2)
	response := RepositoryResponse{}
	queryType := request.Body.Special.Type

	switch queryType {
	case "getFields":
		request.Log("Starting GET-FIELDS sub routine!")
		byteArray := repository.executeGetFields(request)
		response.IsSuccess = true
		response.Message = "Successfully Recieved Field Names"
		response.GetResponseWithBody(byteArray)
	case "getClasses":
		request.Log("Starting GET-CLASSES sub routine")
		byteArray := repository.executeGetClasses(request)
		response.IsSuccess = true
		response.Message = "Successfully Recieved Field Names"
		response.GetResponseWithBody(byteArray)
	case "getNamespaces":
		request.Log("Starting GET-NAMESPACES sub routine")
		byteArray := repository.executeGetNamespaces(request)
		response.IsSuccess = true
		response.Message = "Successfully Recieved Field Names"
		response.GetResponseWithBody(byteArray)
	case "getSelected":
		request.Log("Starting GET-SELECTED sub routine!")
		byteArray := repository.executeGetSelected(request)
		response.IsSuccess = true
		response.Message = "Successfully Recieved Field Names"
		response.GetResponseWithBody(byteArray)

	case "DropClass":
		request.Log("Starting Delete-Class sub routine")
		request.Log("Not implemented in Cloud DataStore repository")
		return getDefaultNotImplemented()
	case "DropNamespace":
		request.Log("Starting Delete-Database sub routine")
		request.Log("Not implemented in Cloud DataStore repository")
		return getDefaultNotImplemented()
	default:
		return repository.GetAll(request)

	}

	return response
}

func (repository GoogleDataStoreRepository) Test(request *messaging.ObjectRequest) {

}

func (repository GoogleDataStoreRepository) getRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
	isGUIDKey := false
	isAutoIncrementId := false //else MANUAL key from the user

	if obj == nil {
		//single request
		if (request.Controls.Id == "-999") || (request.Body.Parameters.AutoIncrement == true) {
			isAutoIncrementId = true
		}

		if (request.Controls.Id == "-888") || (request.Body.Parameters.GUIDKey == true) {
			isGUIDKey = true
		}

	} else {
		//multiple requests
		if (obj[request.Body.Parameters.KeyProperty].(string) == "-999") || (request.Body.Parameters.AutoIncrement == true) {
			isAutoIncrementId = true
		}

		if (obj[request.Body.Parameters.KeyProperty].(string) == "-888") || (request.Body.Parameters.GUIDKey == true) {
			isGUIDKey = true
		}

	}

	if isGUIDKey {
		//GUID Key generation requested!
		returnID = uuid.NewV1().String()
	} else if isAutoIncrementId {
		//Automatic Increment Key generation requested!
		returnID = uuid.NewV1().String()
		ctx := context.Background()
		client, err := repository.getConnection(request)
		ctx = datastore.WithNamespace(ctx, request.Controls.Namespace)

		if err == nil {
			//read from Namespace->domainClassAttributes
			//if there, increment and save.. return id
			//else create new record and save 1.. return 1
			key := datastore.NewKey(ctx, "domainClassAttributes", request.Controls.Class, 0, nil)
			if existingRecord := repository.getByKey(client, ctx, key); existingRecord != nil {
				newId, _ := strconv.Atoi(existingRecord["maxCount"].(string))
				newId++
				//update new Id
				existingRecord["maxCount"] = strconv.Itoa(newId)
				existingRecord["version"] = uuid.NewV1().String()
				//update record
				repository.setAtomicRecord(client, ctx, key, existingRecord)
				returnID = strconv.Itoa(newId)
				return
			} else {
				//No record Available.. Create one.. return 1
				var insertRecord map[string]interface{}
				insertRecord = make(map[string]interface{})
				insertRecord["class"] = request.Controls.Class
				insertRecord["maxCount"] = "1"
				insertRecord["version"] = uuid.NewV1().String()
				repository.setAtomicRecord(client, ctx, key, insertRecord)
				returnID = "1"
				return
			}
		} else {
			returnID = uuid.NewV1().String()
		}
	} else {
		//Manual Key requested!
		if obj == nil {
			returnID = request.Controls.Id
		} else {
			returnID = obj[request.Body.Parameters.KeyProperty].(string)
		}
	}

	return
}

func (repository GoogleDataStoreRepository) getByKey(client *datastore.Client, ctx context.Context, key *datastore.Key) map[string]interface{} {

	var props datastore.PropertyList
	var data map[string]interface{}
	data = make(map[string]interface{})

	if err := client.Get(ctx, key, &props); err != nil {
		term.Write(err.Error(), 1)
		data = nil
	} else {
		for _, value := range props {
			if value.Name != "_os_id" && value.Name != "__osHeaders" {
				data[value.Name] = repository.GQLToGolang(value.Value)
			}
		}
	}

	return data
}

func (repository GoogleDataStoreRepository) setAtomicRecord(client *datastore.Client, ctx context.Context, key *datastore.Key, data map[string]interface{}) {

	var props datastore.PropertyList

	for key, value := range data {
		props = append(props, datastore.Property{Name: key, Value: value})
	}

	_, err := client.Put(ctx, key, &props)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(key)
	}
}

func (repository GoogleDataStoreRepository) GolangToGQL(input interface{}) (value interface{}) {

	varType := reflect.TypeOf(input)

	switch varType.String() {
	case "string":
		value = input
	case "bool":
		value = input
		break
	case "uint":
	case "int":
	case "uint16":
	case "uint32":
	case "uint64":
	case "int8":
	case "int16":
	case "int32":
	case "int64":
		value = input
		break
	case "float32":
	case "float64":
		value = input
		break
	case "byte":
		value = input
		break
	default:
		if byteVal, err := json.Marshal(input); err == nil {
			value = byteVal
		} else {
			value = []byte("{}")
		}
		break
	}

	return
}

func (repository GoogleDataStoreRepository) getSearchToken(input string) (value interface{}) {
	var interfaceType interface{}

	if floatValue, err := strconv.ParseFloat(input, 64); err == nil {
		value = floatValue
		return
	} else if intValue, err := strconv.Atoi(input); err == nil {
		value = intValue
		return
	} else if boolValue, err := strconv.ParseBool(input); err == nil {
		value = boolValue
		return
	} else if err := json.Unmarshal([]byte(input), &interfaceType); err == nil {
		value, _ = json.Marshal(interfaceType)
	} else {
		value = input
		return
	}
	return
}

func (repository GoogleDataStoreRepository) GQLToGolang(input interface{}) (value interface{}) {

	varType := reflect.TypeOf(input)
	switch varType.String() {
	case "string":
		value = input
	case "bool":
		value = input
		break
	case "uint":
	case "int":
	case "uint16":
	case "uint32":
	case "uint64":
	case "int8":
	case "int16":
	case "int32":
	case "int64":
		value = input
		break
	case "float32":
	case "float64":
		value = input
		break
	case "[]byte":
	case "[]uint8":
		var m interface{}
		arr := input.([]byte)
		if string(arr[0]) == "{" || string(arr[0]) == "[" {
			err := json.Unmarshal(input.([]byte), &m)
			if err == nil {
				value = m
			} else {
				term.Write(err.Error(), 1)
				value = input
			}
		} else {
			value = input
		}
		break
	default:
		if byteVal, err := json.Marshal(input); err == nil {
			value = byteVal
		} else {
			value = []byte("{}")
		}
		break
	}

	return
}

func (repository GoogleDataStoreRepository) executeGetFields(request *messaging.ObjectRequest) (returnBytes []byte) {
	ctx := context.Background()
	client, err := repository.getConnection(request)
	ctx = datastore.WithNamespace(ctx, request.Controls.Namespace)

	if err != nil {
		term.Write(err.Error(), 1)
		returnBytes = getEmptyByteObject()
	} else {

		props := make([]datastore.PropertyList, 0)

		var data []string

		var query *datastore.Query

		query = datastore.NewQuery(request.Controls.Class).Limit(1)

		_, err := client.GetAll(ctx, query, &props)
		if err != nil {
			returnBytes = getEmptyByteObject()
			term.Write(err.Error(), 1)
		} else {
			//data recieved! :)
			for index := 0; index < len(props); index++ {
				for _, value := range props[index] {
					if value.Name != "_os_id" && value.Name != "__osHeaders" {
						data = append(data, value.Name)
					}
				}
			}
		}

		returnBytes, _ = json.Marshal(data)
		if len(returnBytes) == 4 || len(returnBytes) == 2 {
			returnBytes = getEmptyByteObject()
		}
	}

	return
}

func (repository GoogleDataStoreRepository) executeGetNamespaces(request *messaging.ObjectRequest) (returnBytes []byte) {
	ctx := context.Background()
	client, err := repository.getConnection(request)
	if err != nil {
		term.Write(err.Error(), 1)
		returnBytes = getEmptyByteObject()
	} else {
		props := make([]datastore.PropertyList, 0)
		var data []string
		query := datastore.NewQuery("__namespace__").Offset(1)
		keys, err := client.GetAll(ctx, query, &props)
		if err != nil {
			returnBytes = getEmptyByteObject()
			term.Write(err.Error(), 1)
		} else {
			fmt.Println(props)
			for index := 0; index < len(keys); index++ {
				data = append(data, keys[index].Name())
			}
		}
		returnBytes, _ = json.Marshal(data)
		if len(returnBytes) == 4 || len(returnBytes) == 2 {
			returnBytes = getEmptyByteObject()
		}
	}
	return
}

func (repository GoogleDataStoreRepository) executeGetClasses(request *messaging.ObjectRequest) (returnBytes []byte) {
	ctx := context.Background()
	client, err := repository.getConnection(request)
	ctx = datastore.WithNamespace(ctx, request.Controls.Namespace)
	if err != nil {
		term.Write(err.Error(), 1)
		returnBytes = getEmptyByteObject()
	} else {
		props := make([]datastore.PropertyList, 0)
		var data []string
		query := datastore.NewQuery("__kind__").Offset(7)
		keys, err := client.GetAll(ctx, query, &props)
		if err != nil {
			returnBytes = getEmptyByteObject()
			term.Write(err.Error(), 1)
		} else {
			fmt.Println(props)
			for index := 0; index < len(keys); index++ {
				data = append(data, keys[index].Name())
			}
		}

		returnBytes, _ = json.Marshal(data)
		if len(returnBytes) == 4 || len(returnBytes) == 2 {
			returnBytes = getEmptyByteObject()
		}
	}
	return
}

func (repository GoogleDataStoreRepository) executeGetSelected(request *messaging.ObjectRequest) (returnBytes []byte) {
	isOrderByAsc := false
	isOrderByDesc := false
	orderbyfield := ""

	skip := 0
	take := 100

	if request.Extras["skip"] != nil {
		if intValue, err := strconv.Atoi(request.Extras["skip"].(string)); err == nil {
			skip = intValue
		}
	}
	if request.Extras["take"] != nil {
		if intValue, err := strconv.Atoi(request.Extras["take"].(string)); err == nil {
			take = intValue
		}
	}
	if request.Extras["orderby"] != nil {
		orderbyfield = request.Extras["orderby"].(string)
		isOrderByAsc = true
	} else if request.Extras["orderbydsc"] != nil {
		orderbyfield = request.Extras["orderbydsc"].(string)
		isOrderByDesc = true
	}

	ctx := context.Background()
	client, err := repository.getConnection(request)
	ctx = datastore.WithNamespace(ctx, request.Controls.Namespace)
	selectedValues := request.Body.Special.Parameters
	if err != nil {
		fmt.Println(err.Error())
		returnBytes = getEmptyByteObject()
	} else {
		props := make([]datastore.PropertyList, 0)
		var data []map[string]interface{}

		var query *datastore.Query

		if isOrderByAsc {
			query = datastore.NewQuery(request.Controls.Class).Offset(skip).Limit(take).Order(orderbyfield)
		} else if isOrderByDesc {
			query = datastore.NewQuery(request.Controls.Class).Offset(skip).Limit(take).Order(("-" + orderbyfield))
		} else {
			query = datastore.NewQuery(request.Controls.Class).Offset(skip).Limit(take)
		}

		_, err := client.GetAll(ctx, query, &props)
		if err != nil {
			term.Write(err.Error(), 1)
		} else {
			//data recieved! :)
			for index := 0; index < len(props); index++ {
				var record map[string]interface{}
				record = make(map[string]interface{})
				for _, value := range props[index] {
					if value.Name != "_os_id" && value.Name != "__osHeaders" {
						if strings.Contains(selectedValues, value.Name) {
							record[value.Name] = repository.GQLToGolang(value.Value)
						}
					}
				}
				data = append(data, record)
			}
		}
		returnBytes, _ = json.Marshal(data)
		if len(returnBytes) == 4 || len(returnBytes) == 2 {
			returnBytes = getEmptyByteObject()
		}
	}
	return
}
