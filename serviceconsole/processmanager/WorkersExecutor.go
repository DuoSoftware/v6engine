package processmanager

//package main

import (
	//"duov6.com/serviceconsole/configuration"
	"duov6.com/serviceconsole/messaging"
	"fmt"
	//"reflect"
	//"encoding/json"
	"github.com/streadway/amqp"
	//"log"
	"strconv"
)

type WorkersExecutor struct {
}

func (w WorkersExecutor) Execute(request *messaging.ServiceRequest) (response messaging.ServiceResponse) {
	response = startAtomicListening(request)
	//fmt.Println("Hola mi amigo!")
	//response.IsSuccess = true
	return response
}

func getWorkers(request *messaging.ServiceRequest) []AbstractWorkers {
	var outWorkers []AbstractWorkers
	//get the number of AbstractWorkers
	outWorkers = make([]AbstractWorkers, getQueueCount(request))
	fmt.Println("Queue Count : " + strconv.Itoa(getQueueCount(request)))
	exchangerMappings := request.Configuration.PublisherConfiguration
	count := 0

	for _, value := range exchangerMappings {
		for key, value2 := range value {

			//generate name for the keyed queues
			tempQueueName := ""
			if value2.Keys == nil {
				tempQueueName = key
				fmt.Println(tempQueueName)
				absWorker := Create(tempQueueName)
				outWorkers[count] = absWorker
				count++
			} else {
				for _, keyValue := range value2.Keys {
					tempQueueName = key + "." + keyValue
					fmt.Println(tempQueueName)
					absWorker := Create(tempQueueName)
					outWorkers[count] = absWorker
					count++
				}
			}
		}
	}

	return outWorkers
}

func getRequiredWorkers(request *messaging.ServiceRequest, selector string) []AbstractWorkers {

	//get required QueueCount
	queuecount := 0
	var junk string //dont care this variable..

	for key, _ := range request.Configuration.PublisherConfiguration {

		for key, value2 := range request.Configuration.PublisherConfiguration[key] {
			fmt.Println(value2)

			if value2.Keys == nil {
				if key == selector {
					queuecount++
				}
			} else {
				for _, keyValue := range value2.Keys {
					if keyValue == selector {
						queuecount++
					}
					junk += keyValue
				}
			}
		}
	}
	//ends here

	var outWorkers []AbstractWorkers
	outWorkers = make([]AbstractWorkers, queuecount)
	fmt.Println("Queue Count : " + strconv.Itoa(queuecount))
	exchangerMappings := request.Configuration.PublisherConfiguration
	count := 0

	for _, value := range exchangerMappings {
		for key, value2 := range value {

			//generate name for the keyed queues
			tempQueueName := ""
			if value2.Keys == nil {
				if key == selector {
					tempQueueName = key
					fmt.Println(tempQueueName)
					absWorker := Create(tempQueueName)
					outWorkers[count] = absWorker
					count++
				}
			} else {
				for _, keyValue := range value2.Keys {
					if keyValue == selector {
						tempQueueName = key + "." + keyValue
						fmt.Println(tempQueueName)
						absWorker := Create(tempQueueName)
						outWorkers[count] = absWorker
						count++
					}
				}
			}
		}
	}

	return outWorkers
}

func getQueueCount(request *messaging.ServiceRequest) (count int) {

	count = 0
	var junk string //dont care this variable..

	for key, _ := range request.Configuration.PublisherConfiguration {

		for _, value2 := range request.Configuration.PublisherConfiguration[key] {
			fmt.Println(value2)

			if value2.Keys == nil {
				count++
			} else {
				arrayIndex := 0
				for _, keyValue := range value2.Keys {
					count++
					junk += keyValue
					arrayIndex++
				}
			}
		}
	}

	return count
}

func startAtomicOperation(request *messaging.ServiceRequest, workerList []AbstractWorkers) (response messaging.ServiceResponse) {

	for _, worker := range workerList {

		fmt.Println("Executing Single Worker : " + worker.GetWorkerName())

		tmpResponse := ExecuteSingleWorker(request, worker)

		if tmpResponse.IsSuccess {
			fmt.Println("Executing worker : " + worker.GetWorkerName() + " - Success")
			response.IsSuccess = true
		} else {
			fmt.Println("Executing worker : " + worker.GetWorkerName() + " - Failed")
			response.IsSuccess = false
		}
	}
	return response
}

func ExecuteSingleWorker(request *messaging.ServiceRequest, worker AbstractWorkers) (response messaging.ServiceResponse) {

	switch request.OperationCode { //GetWorkerName, ExecuteWorker

	case "ExecuteWorker":
		response = worker.ExecuteWorker(request)
	}

	return
}

func startAtomicListening(request *messaging.ServiceRequest) (response messaging.ServiceResponse) {
	fmt.Println("1")
	conn, err := amqp.Dial("amqp://admin:admin@192.168.1.194:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	//create Exchange track array
	var myExchanges []string
	myExchanges = make([]string, len(request.Configuration.PublisherConfiguration))

	noOfProcesses := getQueueCount(request)
	//get no of queues
	fmt.Println("No of Processes : " + strconv.Itoa(noOfProcesses))
	//create msgs array for that.
	var myMsgs []<-chan amqp.Delivery
	myMsgs = make([]<-chan amqp.Delivery, noOfProcesses)

	//create queues

	var myQueues []amqp.Queue
	myQueues = make([]amqp.Queue, noOfProcesses)

	for x := 0; x < noOfProcesses; x++ {
		myQueues[x], err = ch.QueueDeclare(
			"",    // name
			false, // durable
			false, // delete when usused
			true,  // exclusive
			false, // no-wait
			nil,   // arguments
		)
		failOnError(err, "Failed to declare a queue")
	}

	var QueueNameArray []string
	QueueNameArray = make([]string, noOfProcesses+1)
	arrayIndex := 0

	for key, _ := range request.Configuration.PublisherConfiguration {

		myExchanges[arrayIndex] = key

		err = ch.ExchangeDeclare(
			key,      // name
			"direct", // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		)
		failOnError(err, "Failed to declare an exchange")

		for key2, value2 := range request.Configuration.PublisherConfiguration[key] {

			if value2.Keys == nil {
				QueueNameArray[arrayIndex] = key2
				arrayIndex++

			} else {
				for _, keyValue := range value2.Keys {
					fmt.Println(keyValue)
					QueueNameArray[arrayIndex] = key2 + "." + keyValue
					arrayIndex++
				}
			}

		}

		for x := 0; x < noOfProcesses; x++ {
			err = ch.QueueBind(
				myQueues[x].Name,  // queue name
				QueueNameArray[x], // routing key
				key,               // exchange
				false,
				nil)
			failOnError(err, "Failed to bind a queue")
		}

		for x := 0; x < noOfProcesses; x++ {
			myMsgs[x], err = ch.Consume(
				myQueues[x].Name, // queue
				"",               // consumer
				true,             // auto ack
				false,            // exclusive
				false,            // no local
				false,            // no wait
				nil,              // args
			)
			failOnError(err, "Failed to register a consumer")
		}
	}

	fmt.Println(QueueNameArray)
	fmt.Println(myExchanges)

	forever := make(chan bool)

	for x := 0; x < noOfProcesses; x++ {
		ExecuteThread(myMsgs[x], request)
	}

	fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	conn.Close()
	ch.Close()

	return response
}

func ExecuteThread(msgs <-chan amqp.Delivery, request *messaging.ServiceRequest) {
	go func() {
		for d := range msgs {
			//log.Printf("Received a message: %s", d.Body)
			request.Body = d.Body
			convertedWorkers := getRequiredWorkers(request, d.RoutingKey)
			response := startAtomicOperation(request, convertedWorkers)
			fmt.Println(response)

		}

	}()
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Println("Error occured in RabbitMQ! : " + msg)
	}
}
