package repositories

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	//"database/sql"
	//"database/sql/driver"
	//"errors"
	"fmt"
	"github.com/mattbaird/hive"
	"reflect"
	"strconv"
	"strings"
)

type HiveRepository struct {
}

func (repository HiveRepository) GetRepositoryName() string {
	return "Hive DB"
}

func getHiveConnection(request *messaging.ObjectRequest) (conn *hive.HiveConnection, isError bool, errorMessage string) {
	isError = false
	hive.MakePool("192.168.0.97:10000")
	conn, err := hive.GetHiveConn()
	if err != nil {
		isError = true
		errorMessage = err.Error()
		fmt.Println("HIVE connection initilizing failed!")
	} else {
		fmt.Println("HIVE connection initilizing Successful!")
	}

	request.Log("Reusing existing HIVE connection")
	return
}

func (repository HiveRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	conn, isError, errorMessage := getHiveConnection(request)
	if isError == false {
		er, err := conn.Client.Execute("SELECT * FROM " + request.Controls.Class)
		if er == nil && err == nil {

			var myMap map[string]string
			myMap = make(map[string]string)
			recordNumber := 1

			for {
				row, _, _ := conn.Client.FetchOne()
				if row == "" {
					break
				} else {
					var temp []string
					temp = strings.Split(row, "\t")
					var temp2 string
					for i := 0; i < len(temp); i++ {
						temp2 = temp2 + " " + temp[i]
					}
					//fmt.Println(row)
					myMap[strconv.Itoa(recordNumber)] = temp2
					recordNumber = recordNumber + 1
				}
			}
			byteValue, errMarshal := json.Marshal(myMap)

			if errMarshal != nil {
				response.Message = "Conversion to JSON failed!"
				request.Log(response.Message)
			}
			response.IsSuccess = true
			response.GetResponseWithBody(byteValue)
			response.Message = "Successfully retrieved values for all objects in HIVE"
			request.Log(response.Message)
		} else {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in HIVE" + err.Error())
		}
	} else {
		response.GetErrorResponse(errorMessage)
	}
	if conn != nil {
		conn.Checkin()
	}
	return response
}

func (repository HiveRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Get Search not implemented in Hive Db repository")
	return getDefaultNotImplemented()
}

func (repository HiveRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Get Query not implemented in Hive Db repository")
	return getDefaultNotImplemented()
}

func (repository HiveRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Get By Key not implemented in Hive Db repository")
	return getDefaultNotImplemented()
}

func (repository HiveRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	conn, isError, errorMessage := getHiveConnection(request)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		appendKey := request.Controls.Namespace + "." + request.Controls.Class + "."

		noOfQueries := len(request.Body.Objects)
		fmt.Println(noOfQueries)

		for i := 0; i < len(request.Body.Objects); i++ {
			noOfElements := len(request.Body.Objects[i]) - 1

			var keyArray = make([]string, noOfElements)
			var valueArray = make([]string, noOfElements)

			// Process A :start identifying individual data in array and convert to string
			var startIndex int = 0

			for key, value := range request.Body.Objects[i] {

				if key != "__osHeaders" {
					fmt.Println("Key : " + key)

					if str, ok := value.(string); ok {
						fmt.Println(str)
						//Implement all MAP related logic here. All correct data are being caught in here
						if request.Body.Parameters.KeyProperty == key {
							keyArray[startIndex] = key
							valueArray[startIndex] = appendKey + value.(string)
						} else {
							keyArray[startIndex] = key
							valueArray[startIndex] = value.(string)
						}
						startIndex = startIndex + 1

					} else {
						//	fmt.Print("Not String : ")
						fmt.Print(value)
					}
				} else {
					//fmt.Print("Damn __osHeaders Catched! :P")
				}
			}

			var argKeyList string
			var argValueList string

			//Build the query string
			for i := 0; i < noOfElements; i++ {
				if i != noOfElements-1 {
					argKeyList = argKeyList + keyArray[i] + ", "
					argValueList = argValueList + "'" + valueArray[i] + "'" + ", "
				} else {
					argKeyList = argKeyList + keyArray[i]
					argValueList = argValueList + "'" + valueArray[i] + "'"
				}
			}

			//DEBUG USE : Display Query information
			fmt.Println("Table Name : " + request.Controls.Class)
			fmt.Println("Key list : " + argKeyList)
			fmt.Println("Value list : " + argValueList)

			er, err := conn.Client.Execute("insert into table " + request.Controls.Class + " values (" + argValueList + ")")
			if er == nil && err == nil {
				response.IsSuccess = true
				response.Message = "Successfully inserted a single object in to HIVE"
				request.Log(response.Message)
			} else {
				response.IsSuccess = false
				response.GetErrorResponse("Error inserting a single object in to HIVE" + err.Error())
			}

		}

	}

	if conn != nil {
		conn.Checkin()
	}
	return response
}

func (repository HiveRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	conn, isError, errorMessage := getHiveConnection(request)
	fmt.Println(reflect.TypeOf(conn))
	if isError == false {

		appendKey := request.Controls.Namespace + "." + request.Controls.Class + "."
		noOfElements := len(request.Body.Object) - 1

		var keyArray = make([]string, noOfElements)
		var valueArray = make([]string, noOfElements)

		// Process A :start identifying individual data in array and convert to string
		var startIndex int = 0
		for key, value := range request.Body.Object {

			if key != "__osHeaders" {
				fmt.Println("Key : " + key)

				if str, ok := value.(string); ok {
					fmt.Println(str)
					//Implement all MAP related logic here. All correct data are being caught in here
					if request.Body.Parameters.KeyProperty == key {
						keyArray[startIndex] = key
						valueArray[startIndex] = appendKey + value.(string)
					} else {
						keyArray[startIndex] = key
						valueArray[startIndex] = value.(string)
					}
					startIndex = startIndex + 1

				} else {
					//fmt.Print("Not String : ")
					fmt.Print(value)
				}
			} else {
				//fmt.Print("Damn __osHeaders Catched! :P")
			}
		}

		var argKeyList string
		var argValueList string

		//Build the query string
		for i := 0; i < noOfElements; i++ {
			if i != noOfElements-1 {
				argKeyList = argKeyList + keyArray[i] + ", "
				argValueList = argValueList + "'" + valueArray[i] + "'" + ", "
			} else {
				argKeyList = argKeyList + keyArray[i]
				argValueList = argValueList + "'" + valueArray[i] + "'"
			}
		}

		//DEBUG USE : Display Query information
		fmt.Println("Table Name : " + request.Controls.Class)
		fmt.Println("Key list : " + argKeyList)
		fmt.Println("Value list : " + argValueList)

		er, err := conn.Client.Execute("insert into table " + request.Controls.Class + " values (" + argValueList + ")")
		if er == nil && err == nil {
			response.IsSuccess = true
			response.Message = "Successfully inserted a single object in to HIVE"
			request.Log(response.Message)
		} else {
			response.IsSuccess = false
			response.GetErrorResponse("Error inserting a single object in to HIVE" + err.Error())
		}
	} else {
		response.GetErrorResponse(errorMessage)
	}
	if conn != nil {
		conn.Checkin()
	}
	return response
}

func (repository HiveRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	conn, isError, errorMessage := getHiveConnection(request)
	fmt.Println(reflect.TypeOf(conn))
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		fmt.Print("count of objects")
		fmt.Print(len(request.Body.Objects))

		for i := 0; i < len(request.Body.Objects); i++ {
			noOfElements := len(request.Body.Objects[i]) - 2
			var keyUpdate = make([]string, noOfElements)
			var valueUpdate = make([]string, noOfElements)

			var startIndex = 0
			for key, value := range request.Body.Objects[i] {
				if key != request.Body.Parameters.KeyProperty {
					if key != "__osHeaders" {
						keyUpdate[startIndex] = key
						valueUpdate[startIndex] = value.(string)
						startIndex = startIndex + 1
					}
				}
				fmt.Println("Key :" + key)
				fmt.Println(value)

			}

			var argValueList string

			//Build the query string
			for i := 0; i < noOfElements; i++ {
				if i != noOfElements-1 {
					argValueList = argValueList + keyUpdate[i] + " = " + "'" + valueUpdate[i] + "'" + ", "
				} else {
					argValueList = argValueList + keyUpdate[i] + " = " + "'" + valueUpdate[i] + "'"
				}
			}

			//DEBUG USE : Display Query information
			fmt.Println("Table Name : " + request.Controls.Class)
			fmt.Println("Value list : " + argValueList)

			er, err := conn.Client.Execute("UPDATE " + request.Controls.Class + " SET " + argValueList + " WHERE " + request.Body.Parameters.KeyProperty + " =" + "'" + getNoSqlKey(request) + "'")
			//err := collection.Update(bson.M{key: value}, bson.M{"$set": request.Body.Object})
			if er == nil && err == nil {
				response.IsSuccess = true
				response.Message = "Successfully Deleted a single object in to HIVE"
				request.Log(response.Message)
			} else {
				response.IsSuccess = false
				response.GetErrorResponse("Error deleting a single object in to HIVE" + err.Error())
			}
		}

		if conn != nil {
			conn.Checkin()
		}

	}
	return response

	// request.Log("Update Multiple not implemented in Hive Db repository")
	// return getDefaultNotImplemented()
}

func (repository HiveRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	conn, isError, errorMessage := getHiveConnection(request)
	fmt.Println(reflect.TypeOf(conn))
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		noOfElements := len(request.Body.Object) - 2
		var keyUpdate = make([]string, noOfElements)
		var valueUpdate = make([]string, noOfElements)

		var startIndex = 0
		for key, value := range request.Body.Object {
			if key != request.Body.Parameters.KeyProperty {
				if key != "__osHeaders" {
					keyUpdate[startIndex] = key
					valueUpdate[startIndex] = value.(string)
					startIndex = startIndex + 1
				}
			}
			fmt.Println("Key :" + key)
			fmt.Println(value)

		}

		var argValueList string

		//Build the query string
		for i := 0; i < noOfElements; i++ {
			if i != noOfElements-1 {
				argValueList = argValueList + keyUpdate[i] + " = " + "'" + valueUpdate[i] + "'" + ", "
			} else {
				argValueList = argValueList + keyUpdate[i] + " = " + "'" + valueUpdate[i] + "'"
			}
		}

		//DEBUG USE : Display Query information
		fmt.Println("Table Name : " + request.Controls.Class)
		fmt.Println("Value list : " + argValueList)
		fmt.Println("ID to look : " + request.Body.Parameters.KeyProperty)
		fmt.Println("Value for ID : " + getNoSqlKey(request))
		er, err := conn.Client.Execute("UPDATE " + request.Controls.Class + " SET " + argValueList + " WHERE " + request.Body.Parameters.KeyProperty + " =" + "'" + getNoSqlKey(request) + "'")
		//err := collection.Update(bson.M{key: value}, bson.M{"$set": request.Body.Object})
		if er == nil && err == nil {
			response.IsSuccess = true
			response.Message = "Successfully Deleted a single object in to HIVE"
			request.Log(response.Message)
		} else {
			response.IsSuccess = false
			response.GetErrorResponse("Error deleting a single object in to HIVE" + err.Error())
		}
	}
	if conn != nil {
		conn.Checkin()
	}
	return response

	// 	request.Log("Update Single not implemented in Hive Db repository")
	// 	return getDefaultNotImplemented()
}

func (repository HiveRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Delete Multiple not implemented in Hive Db repository : Not Supported by Hive")
	return getDefaultNotImplemented()
}

func (repository HiveRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	conn, isError, errorMessage := getHiveConnection(request)
	fmt.Println(reflect.TypeOf(conn))
	if isError == false {

		appendKey := request.Controls.Namespace + "." + request.Controls.Class + "."
		noOfElements := len(request.Body.Object) - 1

		var keyArray = make([]string, noOfElements)
		var valueArray = make([]string, noOfElements)

		// Process A :start identifying individual data in array and convert to string
		var startIndex int = 0
		for key, value := range request.Body.Object {

			if key != "__osHeaders" {
				fmt.Println("Key : " + key)

				if str, ok := value.(string); ok {
					fmt.Println(str)
					//Implement all MAP related logic here. All correct data are being caught in here
					if request.Body.Parameters.KeyProperty == key {
						keyArray[startIndex] = key
						valueArray[startIndex] = appendKey + value.(string)
					} else {
						keyArray[startIndex] = key
						valueArray[startIndex] = value.(string)
					}
					startIndex = startIndex + 1

				} else {
					//fmt.Print("Not String : ")
					fmt.Print(value)
				}
			} else {
				//fmt.Print("Damn __osHeaders Catched! :P")
			}
		}

		fmt.Println("--------------------------")
		fmt.Println(keyArray[0])
		fmt.Println(valueArray[0])
		fmt.Println("DELETE FROM " + request.Controls.Class + " WHERE " + keyArray[0] + " = " + "'" + valueArray[0] + "'")
		er, err := conn.Client.Execute("DELETE FROM " + request.Controls.Class + " WHERE " + keyArray[0] + " = " + "'" + valueArray[0] + "'")
		if er == nil && err == nil {
			response.IsSuccess = true
			response.Message = "Successfully Deleted a single object in to HIVE"
			request.Log(response.Message)
		} else {
			response.IsSuccess = false
			response.GetErrorResponse("Error deleting a single object in to HIVE")
		}
	} else {
		response.GetErrorResponse(errorMessage)
	}
	if conn != nil {
		conn.Checkin()
	}
	return response

	// request.Log("Delete single not implemented in Hive Db repository : Not Supported by Hive")
	// return getDefaultNotImplemented()
}

func (repository HiveRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Special not implemented in Hive Db repository")
	return getDefaultNotImplemented()
}

func (repository HiveRepository) Test(request *messaging.ObjectRequest) {

}
