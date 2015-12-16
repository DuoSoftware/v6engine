package repositories

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"github.com/twinj/uuid"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/bigtable"
	"io/ioutil"
	"strconv"
	"strings"
)

type GoogleBigTableRepository struct {
}

func (repository GoogleBigTableRepository) GetRepositoryName() string {
	return "GoogleBigTable"
}

func (repository GoogleBigTableRepository) getConnection(request *messaging.ObjectRequest) (client *bigtable.Client, err error) {
	bigTableConfig := request.Configuration.ServerConfiguration["GoogleBigTable"]

	keyFile := bigTableConfig["KeyFile"]
	projectID := bigTableConfig["ProjectID"]
	zone := bigTableConfig["zone"]
	cluster := bigTableConfig["cluster"]

	jsonKey, err := ioutil.ReadFile(keyFile)
	if err != nil {
		term.Write(err.Error(), 1)
	} else {
		conf, err := google.JWTConfigFromJSON(
			jsonKey,
			bigtable.Scope,
		)
		if err != nil {
			term.Write(err.Error(), 1)
		} else {
			ctx := context.Background()
			client, err = bigtable.NewClient(ctx, projectID, zone, cluster, cloud.WithTokenSource(conf.TokenSource(ctx)))
			if err != nil {
				term.Write(err.Error(), 1)
			}
		}
	}

	return
}

func (repository GoogleBigTableRepository) getAdminConnection(request *messaging.ObjectRequest) (client *bigtable.AdminClient, err error) {
	bigTableConfig := request.Configuration.ServerConfiguration["GoogleBigTable"]

	keyFile := bigTableConfig["KeyFile"]
	projectID := bigTableConfig["ProjectID"]
	zone := bigTableConfig["zone"]
	cluster := bigTableConfig["cluster"]

	jsonKey, err := ioutil.ReadFile(keyFile)
	if err != nil {
		term.Write(err.Error(), 1)
	} else {
		conf, err := google.JWTConfigFromJSON(
			jsonKey,
			bigtable.Scope,
			bigtable.AdminScope,
		)
		if err != nil {
			term.Write(err.Error(), 1)
		} else {
			ctx := context.Background()
			client, err = bigtable.NewAdminClient(ctx, projectID, zone, cluster, cloud.WithTokenSource(conf.TokenSource(ctx)))
			if err != nil {
				term.Write(err.Error(), 1)
			}
		}
	}

	return
}

func (repository GoogleBigTableRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-ALL")
	response := RepositoryResponse{}

	ctx := context.Background()
	if client, err := repository.getConnection(request); err == nil {
		tbl := client.Open(request.Controls.Namespace)

		var data []map[string]interface{}

		rowRange := bigtable.PrefixRange((request.Controls.Namespace + "." + request.Controls.Class))
		err := tbl.ReadRows(ctx, rowRange, func(r bigtable.Row) bool {
			for _, v := range r {
				var record map[string]interface{}
				record = make(map[string]interface{})
				for _, o := range v {
					columnTokens := strings.Split(o.Column, ":")
					columnName := columnTokens[1]
					if columnName != "__osHeaders" {
						record[columnName] = repository.GQLToGolang(o.Value)
					}
				}
				data = append(data, record)
			}
			return true
		}, bigtable.RowFilter(bigtable.FamilyFilter(request.Controls.Class)))

		client.Close()

		if err != nil {
			bytesValue := getEmptyByteObject()
			response.IsSuccess = true
			response.Message = "Values Retrieved Successfully from Google BigTable!"
			response.GetResponseWithBody(bytesValue)
		} else {
			response.IsSuccess = true
			response.Message = "Values Retrieved Successfully from Google BigTable!"
			if len(data) > 0 {
				response.GetSuccessResByObject(data)
			} else {
				response.GetResponseWithBody(getEmptyByteObject())
			}
		}

	} else {
		bytesValue := getEmptyByteObject()
		response.IsSuccess = true
		response.Message = "Values Retrieved Successfully from Google BigTable!"
		response.GetResponseWithBody(bytesValue)
	}

	return response
}

func (repository GoogleBigTableRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-Search!", 2)
	response := RepositoryResponse{}

	fieldName := ""
	fieldValue := ""
	if strings.Contains(request.Body.Query.Parameters, ":") {
		tokens := strings.Split(request.Body.Query.Parameters, ":")
		fieldName = tokens[0]
		fieldValue = tokens[1]
		fieldName = strings.TrimSpace(fieldName)
		fieldValue = strings.TrimSpace(fieldValue)
	} else {
		return repository.GetAll(request)
	}

	ctx := context.Background()
	if client, err := repository.getConnection(request); err == nil {
		tbl := client.Open(request.Controls.Namespace)

		var data []map[string]interface{}

		rowRange := bigtable.PrefixRange((request.Controls.Namespace + "." + request.Controls.Class))
		err := tbl.ReadRows(ctx, rowRange, func(r bigtable.Row) bool {
			for _, v := range r {
				var record map[string]interface{}
				record = make(map[string]interface{})
				isValidRecord := false
				for _, o := range v {
					columnTokens := strings.Split(o.Column, ":")
					columnName := columnTokens[1]
					if columnName != "__osHeaders" {
						record[columnName] = repository.GQLToGolang(o.Value)
						if columnName == fieldName && repository.GQLToGolang(o.Value) == repository.getSearchToken(fieldValue) {
							isValidRecord = true
						}
					}
				}

				if isValidRecord {
					data = append(data, record)
					isValidRecord = false
				}
			}
			return true
		}, bigtable.RowFilter(bigtable.FamilyFilter(request.Controls.Class)))

		client.Close()

		if err != nil {
			bytesValue := getEmptyByteObject()
			response.IsSuccess = true
			response.Message = "Values Retrieved Successfully from Google BigTable!"
			response.GetResponseWithBody(bytesValue)
		} else {
			response.IsSuccess = true
			response.Message = "Values Retrieved Successfully from Google BigTable!"
			if len(data) > 0 {
				response.GetSuccessResByObject(data)
			} else {
				response.GetResponseWithBody(getEmptyByteObject())
			}
		}

	} else {
		bytesValue := getEmptyByteObject()
		response.IsSuccess = true
		response.Message = "Values Retrieved Successfully from Google BigTable!"
		response.GetResponseWithBody(bytesValue)
	}

	return response
}

func (repository GoogleBigTableRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-QUERY!")
	response := RepositoryResponse{}
	queryType := request.Body.Query.Type

	switch queryType {
	case "Query":
		if request.Body.Query.Parameters != "*" {
			request.Log("GetQuery not implemented in Google DataStore repository")
			return getDefaultNotImplemented()
		} else {
			return repository.GetAll(request)
		}
	default:
		request.Log(queryType + " is not implemented in Google BigTable repository")
		return getDefaultNotImplemented()
	}
	return response
}

func (repository GoogleBigTableRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-BY-KEY")
	response := RepositoryResponse{}

	ctx := context.Background()
	if client, err := repository.getConnection(request); err == nil {
		tbl := client.Open(request.Controls.Namespace)

		var data []map[string]interface{}

		rowRange := bigtable.SingleRow(getNoSqlKey(request))
		err := tbl.ReadRows(ctx, rowRange, func(r bigtable.Row) bool {
			for _, v := range r {
				var record map[string]interface{}
				record = make(map[string]interface{})
				for _, o := range v {
					columnTokens := strings.Split(o.Column, ":")
					columnName := columnTokens[1]
					if columnName != "__osHeaders" {
						record[columnName] = repository.GQLToGolang(o.Value)
					}
				}
				data = append(data, record)
			}
			return true
		}, bigtable.RowFilter(bigtable.FamilyFilter(request.Controls.Class)))

		client.Close()

		if err != nil {
			bytesValue := getEmptyByteObject()
			response.IsSuccess = true
			response.Message = "Values Retrieved Successfully from Google BigTable!"
			response.GetResponseWithBody(bytesValue)
		} else {
			response.IsSuccess = true
			response.Message = "Values Retrieved Successfully from Google BigTable!"
			if len(data) > 0 {
				response.GetSuccessResByObject(data)
			} else {
				response.GetResponseWithBody(getEmptyByteObject())
			}
		}

	} else {
		bytesValue := getEmptyByteObject()
		response.IsSuccess = true
		response.Message = "Values Retrieved Successfully from Google BigTable!"
		response.GetResponseWithBody(bytesValue)
	}

	return response
}

func (repository GoogleBigTableRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-MULTIPLE")
	return repository.setManyBigTable(request)
}

func (repository GoogleBigTableRepository) setManyBigTable(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	var idData map[string]interface{}
	idData = make(map[string]interface{})

	ctx := context.Background()
	client, err := repository.getConnection(request)

	if err != nil {
		response.IsSuccess = false
		response.Message = "Connection Failed"
		return response
	}

	insertStatus := make([]bool, len(request.Body.Objects))

	//validate schema
	repository.validateSchema(request)

	//open table
	tbl := client.Open(request.Controls.Namespace)

	for index, obj := range request.Body.Objects {
		id := repository.getRecordID(request, obj)
		idData[strconv.Itoa(index)] = id
		request.Body.Objects[index][request.Body.Parameters.KeyProperty] = id

		//check if record available in DB
		if temp := repository.getByKey(tbl, ctx, request, getNoSqlKeyById(request, obj)); temp != nil {
			//delete row
			mut := bigtable.NewMutation()
			mut.DeleteRow()
			err := tbl.Apply(ctx, getNoSqlKeyById(request, obj), mut)
			if err != nil {
				request.Log(err.Error())
			}
		}

		//Insert New Record

		mut := bigtable.NewMutation()

		for key, value := range obj {
			mut.Set(request.Controls.Class, key, bigtable.Now(), getByteByValue(value))
		}

		err = tbl.Apply(ctx, getNoSqlKeyById(request, obj), mut)
		if err != nil {
			insertStatus[index] = false
		} else {
			insertStatus[index] = true
		}

	}

	isAllDone := true
	for _, status := range insertStatus {
		if !status {
			isAllDone = false
			break
		}
	}
	fmt.Println(insertStatus)
	fmt.Println(isAllDone)
	if !isAllDone {
		response.IsSuccess = false
		response.Message = "Inserting Some Elements Failed"
	} else {
		response.IsSuccess = true
		response.Message = "Success inserting/updating multiple values in BigTable"
	}

	client.Close()
	var DataMap []map[string]interface{}
	DataMap = make([]map[string]interface{}, 1)
	var idMap map[string]interface{}
	idMap = make(map[string]interface{})
	idMap["ID"] = idData
	DataMap[0] = idMap
	response.Data = DataMap

	return response
}

func (repository GoogleBigTableRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-SINGLE")
	return repository.setOneBigTable(request)
}

func (repository GoogleBigTableRepository) setOneBigTable(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	id := repository.getRecordID(request, request.Body.Object)
	request.Controls.Id = id
	request.Body.Object[request.Body.Parameters.KeyProperty] = id

	//validate schema
	repository.validateSchema(request)

	ctx := context.Background()
	client, err := repository.getConnection(request)

	if err != nil {
		response.IsSuccess = false
		response.Message = "Connection Failed"
		return response
	}

	//open table
	tbl := client.Open(request.Controls.Namespace)

	//check if record available in DB
	if temp := repository.getByKey(tbl, ctx, request, getNoSqlKey(request)); temp != nil {
		//delete row
		mut := bigtable.NewMutation()
		mut.DeleteRow()
		err := tbl.Apply(ctx, getNoSqlKey(request), mut)
		if err != nil {
			response.IsSuccess = false
			response.Message = err.Error()
			return response
		}
	}

	//Insert New Record
	mut := bigtable.NewMutation()

	for key, value := range request.Body.Object {
		mut.Set(request.Controls.Class, key, bigtable.Now(), getByteByValue(value))
	}

	err = tbl.Apply(ctx, getNoSqlKey(request), mut)
	if err != nil {
		response.IsSuccess = false
		response.Message = err.Error()
	} else {
		response.IsSuccess = true
		response.Message = "Successfully Inserted/Update BigTable"
	}

	client.Close()
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

func (repository GoogleBigTableRepository) validateSchema(request *messaging.ObjectRequest) {
	ctx := context.Background()
	adminClient, _ := repository.getAdminConnection(request)

	//validate Namespace
	tableNames, err := adminClient.Tables(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(tableNames)

	tableString := ""

	for _, name := range tableNames {
		tableString += name + "|"
	}

	if !strings.Contains(tableString, request.Controls.Namespace) {
		fmt.Println("Creating table..")
		err = adminClient.CreateTable(ctx, request.Controls.Namespace)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	//validate Class
	err2 := adminClient.CreateColumnFamily(ctx, request.Controls.Namespace, request.Controls.Class)
	if err != nil {
		fmt.Println(err2.Error())
	} else {
		fmt.Println("Created New Class : " + request.Controls.Class)
	}

	adminClient.Close()

}

func (repository GoogleBigTableRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-MULTIPLE")
	return repository.setManyBigTable(request)
}

func (repository GoogleBigTableRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-SINGLE")
	return repository.setOneBigTable(request)
}

func (repository GoogleBigTableRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-MULTIPLE")
	response := RepositoryResponse{}

	ctx := context.Background()
	client, err := repository.getConnection(request)
	if err == nil {
		tbl := client.Open(request.Controls.Namespace)
		for _, obj := range request.Body.Objects {
			mut := bigtable.NewMutation()
			mut.DeleteRow()
			err := tbl.Apply(ctx, getNoSqlKeyById(request, obj), mut)
			if err != nil {
				response.IsSuccess = false
				response.Message = "No Connection to BigTable : " + err.Error()
				return response
			}
		}
		client.Close()
	} else {
		response.IsSuccess = false
		response.Message = "No Connection to BigTable : " + err.Error()
		request.Log(err.Error())
	}

	return response
}

func (repository GoogleBigTableRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-SINGLE")
	response := RepositoryResponse{}

	ctx := context.Background()
	client, err := repository.getConnection(request)
	if err == nil {
		tbl := client.Open(request.Controls.Namespace)
		mut := bigtable.NewMutation()
		mut.DeleteRow()
		err := tbl.Apply(ctx, getNoSqlKey(request), mut)
		if err != nil {
			response.IsSuccess = false
			response.Message = "No Connection to BigTable : " + err.Error()
			return response
		}
	} else {
		response.IsSuccess = false
		response.Message = "No Connection to BigTable : " + err.Error()
		request.Log(err.Error())
	}
	return response
}

func (repository GoogleBigTableRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
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
		if status := repository.executeDropClass(request); status {
			response.IsSuccess = true
			response.Message = "Successfully dropped class : " + request.Controls.Class
		} else {
			response.IsSuccess = false
			response.Message = "Failed dropping class : " + request.Controls.Class
		}
	case "DropNamespace":
		request.Log("Starting Delete-Database sub routine")
		if status := repository.executeDropNamespace(request); status {
			response.IsSuccess = true
			response.Message = "Successfully dropped class : " + request.Controls.Class
		} else {
			response.IsSuccess = false
			response.Message = "Failed dropping class : " + request.Controls.Class
		}
	default:
		return repository.GetAll(request)

	}

	return response
}

func (repository GoogleBigTableRepository) Test(request *messaging.ObjectRequest) {

}

func (repository GoogleBigTableRepository) getRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
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
		/*//Automatic Increment Key generation requested!
		returnID = uuid.NewV1().String()
		ctx := context.Background()
		client, err := repository.getConnection(request)

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
		} else {*/
		returnID = uuid.NewV1().String()
		//	}
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

func (repository GoogleBigTableRepository) getByKey(tbl *bigtable.Table, ctx context.Context, request *messaging.ObjectRequest, key string) (data map[string]interface{}) {
	data = make(map[string]interface{})

	rowRange := bigtable.SingleRow(key)
	err := tbl.ReadRows(ctx, rowRange, func(r bigtable.Row) bool {
		for _, v := range r {
			for _, o := range v {
				columnTokens := strings.Split(o.Column, ":")
				columnName := columnTokens[1]
				if columnName != "__osHeaders" {
					data[columnName] = repository.GQLToGolang(o.Value)
				}
			}
		}
		return true
	}, bigtable.RowFilter(bigtable.FamilyFilter(request.Controls.Class)))
	fmt.Println(data)
	if err != nil {
		data = nil
	}

	fmt.Println(data)

	return

}

func (repository GoogleBigTableRepository) setAtomicRecord(client *bigtable.Client, ctx context.Context, key string, data map[string]interface{}) {

}

func (repository GoogleBigTableRepository) getSearchToken(input string) (value interface{}) {
	var interfaceType interface{}

	if intValue, err := strconv.Atoi(input); err == nil {
		value = intValue
		return
	} else if floatValue, err := strconv.ParseFloat(input, 32); err == nil {
		value = floatValue
		return
	} else if floatValue, err := strconv.ParseFloat(input, 64); err == nil {
		value = floatValue
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

func (repository GoogleBigTableRepository) GQLToGolang(input []byte) (value interface{}) {

	var boolValue bool
	var intValue int
	var floatValue64 float64
	var floatValue32 float32
	var stringValue string
	var interfaceValue interface{}

	if err := json.Unmarshal(input, &boolValue); err == nil {
		value = boolValue
	} else if err := json.Unmarshal(input, &intValue); err == nil {
		value = intValue
	} else if err := json.Unmarshal(input, &floatValue32); err == nil {
		value = floatValue32
	} else if err := json.Unmarshal(input, &floatValue64); err == nil {
		value = floatValue64
	} else if err := json.Unmarshal(input, &stringValue); err == nil {
		value = stringValue
	} else if err := json.Unmarshal(input, &interfaceValue); err == nil {
		value = interfaceValue
	} else {
		value = input
	}

	return
}

func (repository GoogleBigTableRepository) executeGetFields(request *messaging.ObjectRequest) (returnBytes []byte) {
	ctx := context.Background()
	var fieldNames []string

	if client, err := repository.getConnection(request); err == nil {
		tbl := client.Open(request.Controls.Namespace)

		rowRange := bigtable.PrefixRange((request.Controls.Namespace + "." + request.Controls.Class))
		_ = tbl.ReadRows(ctx, rowRange, func(r bigtable.Row) bool {
			for _, v := range r {
				for _, o := range v {
					columnTokens := strings.Split(o.Column, ":")
					columnName := columnTokens[1]
					if columnName != "__osHeaders" {
						fieldNames = append(fieldNames, columnName)
					}
				}
				break
			}
			return true
		}, bigtable.RowFilter(bigtable.FamilyFilter(request.Controls.Class)), bigtable.LimitRows(1))

		client.Close()
	}

	returnBytes, _ = json.Marshal(fieldNames)
	return
}

func (repository GoogleBigTableRepository) executeGetNamespaces(request *messaging.ObjectRequest) (returnBytes []byte) {
	ctx := context.Background()
	adminClient, _ := repository.getAdminConnection(request)

	tableNames, err := adminClient.Tables(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	} else {
		returnBytes, _ = json.Marshal(tableNames)
	}
	adminClient.Close()
	return
}

func (repository GoogleBigTableRepository) executeGetClasses(request *messaging.ObjectRequest) (returnBytes []byte) {
	ctx := context.Background()
	adminClient, _ := repository.getAdminConnection(request)

	tableInforStructs, err := adminClient.TableInfo(ctx, request.Controls.Namespace)
	if err != nil {
		fmt.Println(err.Error())
		return
	} else {
		returnBytes, _ = json.Marshal(tableInforStructs.Families)
	}
	adminClient.Close()
	return
}

func (repository GoogleBigTableRepository) executeGetSelected(request *messaging.ObjectRequest) (returnBytes []byte) {

	selectedValues := strings.Split((strings.TrimSpace(request.Body.Special.Parameters)), " ")
	selectedValuesString := ""

	for _, value := range selectedValues {
		selectedValuesString += "|" + value
	}

	ctx := context.Background()
	if client, err := repository.getConnection(request); err == nil {
		tbl := client.Open(request.Controls.Namespace)

		var data []map[string]interface{}

		rowRange := bigtable.PrefixRange((request.Controls.Namespace + "." + request.Controls.Class))
		_ = tbl.ReadRows(ctx, rowRange, func(r bigtable.Row) bool {
			for _, v := range r {
				var record map[string]interface{}
				record = make(map[string]interface{})
				for _, o := range v {
					columnTokens := strings.Split(o.Column, ":")
					columnName := columnTokens[1]
					if strings.Contains(selectedValuesString, columnName) {
						record[columnName] = repository.GQLToGolang(o.Value)
					}
				}
				data = append(data, record)
			}
			return true
		}, bigtable.RowFilter(bigtable.FamilyFilter(request.Controls.Class)))

		client.Close()

		returnBytes, _ = json.Marshal(data)
	} else {
		returnBytes = getEmptyByteObject()
	}
	return
}

func (repository GoogleBigTableRepository) executeDropNamespace(request *messaging.ObjectRequest) (status bool) {
	ctx := context.Background()
	status = true

	if adminClient, err := repository.getAdminConnection(request); err == nil {
		if err = adminClient.DeleteTable(ctx, request.Controls.Namespace); err != nil {
			status = false
		}
		adminClient.Close()
	} else {
		status = false
	}
	return
}

func (repository GoogleBigTableRepository) executeDropClass(request *messaging.ObjectRequest) (status bool) {
	ctx := context.Background()
	status = true

	if adminClient, err := repository.getAdminConnection(request); err == nil {
		if err = adminClient.DeleteColumnFamily(ctx, request.Controls.Namespace, request.Controls.Class); err != nil {
			status = false
		}
		adminClient.Close()
	} else {
		status = false
	}
	return
}
