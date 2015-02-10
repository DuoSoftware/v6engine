package repositories

//package main

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	//"log"
	"reflect"
	//"strconv"
	"strings"
)

type CassandraRepository struct {
}

func (repository CassandraRepository) GetRepositoryName() string {
	return "Cassandra DB"
}

func getCassandraConnection(request *messaging.ObjectRequest) (session *gocql.Session, isError bool, errorMessage string) {

	//creating KeySpace out of namespace
	var temp []string
	temp = strings.Split(request.Controls.Namespace, ".")
	//end
	isError = false
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = temp[1]
	session, err := cluster.CreateSession()
	if err != nil {
		isError = false
		errorMessage = err.Error()
		request.Log("Cassandra connection initilizing failed!")
	}
	//defer session.Close()
	fmt.Println(reflect.TypeOf(session))

	request.Log("Reusing existing Cassandra connection")
	return
}

func (repository CassandraRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	session, isError, errorMessage := getCassandraConnection(request)
	fmt.Println(reflect.TypeOf(session))
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		isError = false
		//Process A : Get Count of DB

		var myMap map[string]interface{}
		myMap = make(map[string]interface{})

		iter := session.Query("SELECT COUNT(*) FROM " + request.Controls.Class).Iter()
		for iter.MapScan(myMap) {
			for key, value := range myMap {
				myMap[key] = value

			}
		}
		err := iter.Close()
		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in Cassandra" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully retrieved values for all objects in Cassandra"
			request.Log(response.Message)
		}

		var count int64

		for k, v := range myMap {
			fmt.Print("Key : " + k)
			fmt.Println(v)
			count = v.(int64)
		}

		//PROCESS A Ends here

		var myMap2 []map[string]interface{}
		myMap2 = make(([]map[string]interface{}), count)

		var n int64
		for n = 0; n < count; n++ {
			myMap2[n] = make(map[string]interface{})

		}

		iter2 := session.Query("SELECT * FROM " + request.Controls.Class).Iter()

		my, isErr := iter2.SliceMap()

		if isErr != nil {
			//Handle Error!
		} else {
			for n = 0; n < count; n++ {
				for key, value := range myMap2[n] {
					myMap2[n][key] = value.(string)

				}
			}
		}

		iter2.Close()

		byteValue, errMarshal := json.Marshal(my)
		if errMarshal != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in Cassandra" + err.Error())
		} else {
			response.IsSuccess = true
			response.GetResponseWithBody(byteValue)
			response.Message = "Successfully retrieved values for all objects in mongo"
			request.Log(response.Message)
		}
	}

	return response
}

func (repository CassandraRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Get Search not implemented in Cassandra Db repository")
	return getDefaultNotImplemented()
}

func (repository CassandraRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Get Query not implemented in Casssandra Db repository")
	return getDefaultNotImplemented()
}

func (repository CassandraRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	session, isError, errorMessage := getCassandraConnection(request)
	fmt.Println(reflect.TypeOf(session))
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		isError = false

		fmt.Println("Id key : " + request.Controls.Id)

		var myMap map[string]interface{}
		myMap = make(map[string]interface{})

		iter := session.Query("SELECT * FROM " + request.Controls.Class + " where Id = '" + getNoSqlKey(request) + "'").Iter()
		for iter.MapScan(myMap) {
			for key, value := range myMap {
				myMap[key] = value.(string)

			}
		}
		err := iter.Close()
		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in Cassandra" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully retrieved values for all objects in Cassandra"
			request.Log(response.Message)
		}

		for k, v := range myMap {
			fmt.Print("Key : " + k)
			fmt.Println("value : " + v.(string))
		}

		byteValue, errMarshal := json.Marshal(myMap)
		if errMarshal != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in Cassandra" + err.Error())
		} else {
			response.IsSuccess = true
			response.GetResponseWithBody(byteValue)
			response.Message = "Successfully retrieved values for all objects in mongo"
			request.Log(response.Message)
		}
	}

	return response
}

func (repository CassandraRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	session, isError, errorMessage := getCassandraConnection(request)
	fmt.Println(reflect.TypeOf(session))
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
						fmt.Print("Not String : ")
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

			err := session.Query("INSERT INTO " + request.Controls.Class + " (" + argKeyList + ") VALUES (" + argValueList + ")").Exec()
			if err != nil {
				response.IsSuccess = false
				response.GetErrorResponse("Error inserting one object in Cassandra" + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully inserted one object in Cassandra"
				request.Log(response.Message)
			}

		}

	}
	return response
}

func (repository CassandraRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	session, isError, errorMessage := getCassandraConnection(request)
	fmt.Println(reflect.TypeOf(session))
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

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

		err := session.Query("INSERT INTO " + request.Controls.Class + " (" + argKeyList + ") VALUES (" + argValueList + ")").Exec()
		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error inserting one object in Cassandra" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted one object in Cassandra"
			request.Log(response.Message)
		}
	}
	return response
}

func (repository CassandraRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	session, isError, errorMessage := getCassandraConnection(request)
	fmt.Println(reflect.TypeOf(session))
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

			err := session.Query("UPDATE " + request.Controls.Class + " SET " + argValueList + " WHERE " + request.Body.Parameters.KeyProperty + " =" + "'" + getNoSqlKey(request) + "'").Exec()
			//err := collection.Update(bson.M{key: value}, bson.M{"$set": request.Body.Object})
			if err != nil {
				response.IsSuccess = false
				request.Log("Error updating object in Cassandra  : " + getNoSqlKey(request) + ", " + err.Error())
				response.GetErrorResponse("Error updating one object in Cassandra because no match was found!" + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully updating one object in Cassandra "
				request.Log(response.Message)
			}
		}

	}
	return response
}

func (repository CassandraRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	session, isError, errorMessage := getCassandraConnection(request)
	fmt.Println(reflect.TypeOf(session))
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

		err := session.Query("UPDATE " + request.Controls.Class + " SET " + argValueList + " WHERE " + request.Body.Parameters.KeyProperty + " =" + "'" + getNoSqlKey(request) + "'").Exec()
		//err := collection.Update(bson.M{key: value}, bson.M{"$set": request.Body.Object})
		if err != nil {
			response.IsSuccess = false
			request.Log("Error updating object in Cassandra  : " + getNoSqlKey(request) + ", " + err.Error())
			response.GetErrorResponse("Error updating one object in Cassandra because no match was found!" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully updating one object in Cassandra "
			request.Log(response.Message)
		}

	}
	return response
}

func (repository CassandraRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Delete Multiple not implemented in Cassandra Db repository : Not Supported by Cassandra")
	return getDefaultNotImplemented()
}

func (repository CassandraRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	session, isError, errorMessage := getCassandraConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		err := session.Query("DELETE FROM " + request.Controls.Class + " WHERE " + request.Body.Parameters.KeyProperty + " = '" + getNoSqlKey(request) + "'").Exec()
		if err != nil {
			response.IsSuccess = false
			request.Log("Error deleting object in Cassandra  : " + err.Error())
			response.GetErrorResponse("Error deleting one object in Cassandra because no match was found!" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully deleted one object in Cassandra"
			request.Log(response.Message)
		}
	}

	return response
}

func (repository CassandraRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Special not implemented in Cassandra Db repository")
	return getDefaultNotImplemented()
}

func (repository CassandraRepository) Test(request *messaging.ObjectRequest) {

}
