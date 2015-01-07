package storageengines

import (
	//"bytes"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	//"reflect"
)

type QueuedStorageEngine struct {
}

func failOnError(err error, msg string) {
	fmt.Println("RABBITMQ")
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func (r QueuedStorageEngine) Store(request *messaging.ObjectRequest) (response repositories.RepositoryResponse) {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"publisher_01", // name
		"direct",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	body, errMarshal := json.Marshal(request)

	if errMarshal != nil {
		response.Message = "Conversion to JSON failed!"
		request.Log(response.Message)
	} else {
		response.Message = "Conversion to JSON Success!"
		request.Log(response.Message)
		log.Printf(" [x] Sent %s", body)

	}
	err = ch.Publish(
		"publisher_01", // exchange
		"Queued",       // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			// ContentType: "text/plain",
			ContentType: "*messaging.ObjectRequest",
			Body:        body,
		})
	failOnError(err, "Failed to publish a message")
	if err == nil {
		response.IsSuccess = true
	}

	//log.Printf(" [x] Sent %s", body)

	return response

}

//............. BACKUP CODE...................

// package storageengines

// import (
// 	"bytes"
// 	"duov6.com/objectstore/messaging"
// 	"duov6.com/objectstore/repositories"
// 	"encoding/json"
// 	"fmt"
// 	"github.com/streadway/amqp"
// 	"log"
// 	"reflect"
// )

// type QueuedStorageEngine struct {
// }

// type QueueData struct {
// 	request        *messaging.ObjectRequest
// 	repositoryList []repositories.AbstractRepository
// 	successAction  int
// 	failAction     int
// }

// func (q QueuedStorageEngine) Store(request *messaging.ObjectRequest) (response repositories.RepositoryResponse) {

// 	//1 = COMMIT, 2 = ROLLBACK, 3 = BREAK
// 	var successAction int = 0
// 	var failAction int = 0
// 	var engineMappings map[string]string

// 	switch request.Controls.Operation { //CREATE, READ, UPDATE, DELETE, SPECIAL
// 	case "insert":
// 		successAction = 1
// 		failAction = 2
// 		if request.Controls.Multiplicity == "single" {
// 			request.Log("Getting settings for single insert")
// 			engineMappings = request.Configuration.StoreConfiguration["INSERT-SINGLE"]
// 		} else {
// 			request.Log("Getting settings for multiple insert")
// 			engineMappings = request.Configuration.StoreConfiguration["INSERT-MULTIPLE"]
// 		}
// 	case "read-all":
// 		successAction = 3
// 		failAction = 1
// 		request.Log("Getting settings for get all")
// 		engineMappings = request.Configuration.StoreConfiguration["GET-ALL"]
// 	case "read-key":
// 		successAction = 3
// 		failAction = 1
// 		request.Log("Getting settings for get by key")
// 		engineMappings = request.Configuration.StoreConfiguration["GET-KEY"]
// 	case "read-keyword":
// 		successAction = 3
// 		failAction = 1
// 		request.Log("Getting settings for get by keyword")
// 		engineMappings = request.Configuration.StoreConfiguration["GET-QUERY"]
// 	case "read-filter":
// 		successAction = 3
// 		failAction = 1
// 		request.Log("Getting settings for get by filtering")
// 		engineMappings = request.Configuration.StoreConfiguration["GET-SEARCH"]

// 	case "update":
// 		successAction = 1
// 		failAction = 2
// 		if request.Controls.Multiplicity == "single" {
// 			request.Log("Getting settings for single update")
// 			engineMappings = request.Configuration.StoreConfiguration["UPDATE-SINGLE"]
// 		} else {
// 			request.Log("Getting settings for multiple update")
// 			engineMappings = request.Configuration.StoreConfiguration["UPDATE-MULTIPLE"]
// 		}
// 	case "delete":
// 		successAction = 1
// 		failAction = 2
// 		if request.Controls.Multiplicity == "single" {
// 			request.Log("Getting settings for single delete")
// 			engineMappings = request.Configuration.StoreConfiguration["DELETE-SINGLE"]
// 		} else {
// 			request.Log("Getting settings for multiple delete")
// 			engineMappings = request.Configuration.StoreConfiguration["DELETE-MULTIPLE"]
// 		}
// 	case "special":
// 		successAction = 3
// 		failAction = 1
// 		request.Log("Getting settings for special operation")
// 		engineMappings = request.Configuration.StoreConfiguration["SPECIAL"]

// 	}

// 	convertedRepositories := getQueuedRepositories(engineMappings)

// 	dataInputStruct := QueueData{request, convertedRepositories, successAction, failAction}

// 	inputByteValue, errMarshal := json.Marshal(dataInputStruct)
// 	if errMarshal != nil {
// 		response.GetErrorResponse("Error converting Object response to JSON format")
// 	} else {
// 		//execute PRODUCER
// 		RabbitProducer(inputByteValue)
// 	}

// 	response = startQueuedAtomicOperation(request, convertedRepositories, successAction, failAction)

// 	return
// }

// func getQueuedRepositories(engineMappings map[string]string) []repositories.AbstractRepository {
// 	var outRepositories []repositories.AbstractRepository

// 	outRepositories = make([]repositories.AbstractRepository, len(engineMappings))

// 	count := -1

// 	for _, v := range engineMappings {
// 		count++
// 		absRepository := repositories.Create(v)
// 		outRepositories[count] = absRepository
// 	}

// 	return outRepositories
// }

// func startQueuedAtomicOperation(request *messaging.ObjectRequest, repositoryList []repositories.AbstractRepository, successAction int, failAction int) (response repositories.RepositoryResponse) {
// 	dataInputStruct := QueueData{request, repositoryList, successAction, failAction}
// 	currentInputByteValue, errMarshal := json.Marshal(dataInputStruct)
// 	if errMarshal != nil {
// 		response.GetErrorResponse("Error converting Object response to JSON format")
// 	}

// 	outputByteValue := RabbitWoker()
// 	fmt.Println("Got output")

// 	if bytes.Equal(currentInputByteValue, outputByteValue) {
// 		if reflect.DeepEqual(currentInputByteValue, outputByteValue) {

// 			canRollback := false
// 			for _, repository := range repositoryList {

// 				request.Log("Executing repository : " + repository.GetRepositoryName())

// 				tmpResponse := repositories.Execute(request, repository)
// 				canBreak := false

// 				if tmpResponse.IsSuccess {
// 					request.Log("Executing repository : " + repository.GetRepositoryName() + " - Success")
// 					switch successAction {
// 					case 1:
// 						response = tmpResponse
// 						continue
// 					case 3:
// 						response = tmpResponse
// 						canBreak = true
// 					}
// 				} else {
// 					request.Log("Executing repository : " + repository.GetRepositoryName() + " - Failed")
// 					switch failAction {
// 					case 1:
// 						continue
// 					case 2:
// 						canRollback = true
// 						canBreak = true
// 					case 3:
// 						response = tmpResponse
// 						canBreak = true
// 					}
// 				}

// 				if canBreak == true {
// 					break
// 				}

// 				//1 = COMMIT, 2 = ROLLBACK, 3 = BREAK

// 			}

// 			if canRollback {
// 				request.Log("Transaction failed Rollbacking!!!")
// 			}
// 		} else {
// 			fmt.Println("Error matching byte array")
// 		}

// 	} else {
// 		fmt.Println("Not Equal byte arrays!")
// 	}
// 	return
// }

// func RabbitProducer(inputByteValue []byte) {
// 	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
// 	failOnError(err, "Failed to connect to RabbitMQ")
// 	defer conn.Close()

// 	ch, err := conn.Channel()
// 	failOnError(err, "Failed to open a channel")
// 	defer ch.Close()

// 	body := inputByteValue
// 	err = ch.Publish(
// 		"",           // exchange
// 		"task_queue", // routing key
// 		false,        // mandatory
// 		false,        // Immediage
// 		amqp.Publishing{
// 			DeliveryMode: amqp.Persistent,
// 			ContentType:  "text/plain",
// 			Body:         []byte(body),
// 		})
// 	failOnError(err, "Failed to publish a message")

// }

// func RabbitWoker() (outputByteValue []byte) {
// 	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
// 	failOnError(err, "Failed to connect to RabbitMQ")
// 	defer conn.Close()

// 	ch, err := conn.Channel()
// 	failOnError(err, "Failed to open a channel")
// 	defer ch.Close()

// 	q, err := ch.QueueDeclare(
// 		"task_queue", // name
// 		true,         // durable
// 		false,        // delete when unused
// 		false,        // exclusive
// 		false,        // no-wait
// 		nil,          // arguments
// 	)
// 	failOnError(err, "Failed to declare a queue")

// 	err = ch.Qos(
// 		1,     // prefetch count
// 		0,     // prefetch size
// 		false, // global
// 	)
// 	failOnError(err, "Failed to set QoS")

// 	msgs, err := ch.Consume(
// 		q.Name, // queue
// 		"",     // consumer
// 		false,  // auto-ack
// 		false,  // exclusive
// 		false,  // no-local
// 		false,  // no-wait
// 		nil,    // args
// 	)
// 	failOnError(err, "Failed to register a consumer")

// 	var tempByteArray []byte

// 	for d := range msgs {
// 		fmt.Println("Received a message: %s", d.Body)
// 		d.Ack(false)
// 		tempByteArray = d.Body
// 		break

// 	}

// 	outputByteValue = tempByteArray
// 	return outputByteValue
// }

// func failOnError(err error, msg string) {
// 	if err != nil {
// 		log.Fatalf("%s: %s", msg, err)
// 		panic(fmt.Sprintf("%s: %s", msg, err))
// 	}
// }
