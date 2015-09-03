package processmanager

import (
	"duov6.com/serviceconsole/messaging"
	//"encoding/json"
	"duov6.com/serviceconsole/configuration"
	"fmt"
	"github.com/streadway/amqp"
	//"log"
	"strconv"
)

type WorkersExecutor struct {
}

func (w WorkersExecutor) Execute(request *messaging.ServiceRequest) (response messaging.ServiceResponse) {
	request.Log("Starting new Service Console Session!")
	fmt.Println("Starting new Service Console Session!")
	var initialSlice []string
	initialSlice = make([]string, 0)
	request.MessageStack = initialSlice

	response = startAtomicListening(request)

	return response
}

func getRequiredWorkers(request *messaging.ServiceRequest, selector string) []AbstractWorkers {

	//get required QueueCount
	queuecount := 0
	var junk string //dont care this variable..

	for key, _ := range request.Configuration.PublisherConfiguration {
		//	fmt.Println(key)
		for key, value2 := range request.Configuration.PublisherConfiguration[key] {
			//	fmt.Println(key)
			//	fmt.Println(value2)
			if value2.Keys == nil {
				if key == selector {
					queuecount++

				}
			} else {
				for _, keyValue := range value2.Keys {

					if (key + "." + keyValue) == selector {
						queuecount++
					}
					junk += keyValue
				}
			}
		}
	}
	//ends here

	fmt.Print("Needed Worker Count : ")
	fmt.Println(queuecount)

	var outWorkers []AbstractWorkers
	outWorkers = make([]AbstractWorkers, queuecount)
	//fmt.Println("Queue Count : " + strconv.Itoa(queuecount))
	exchangerMappings := request.Configuration.PublisherConfiguration
	count := 0

	for _, value := range exchangerMappings {
		for key, value2 := range value {
			//fmt.Println(key)
			//fmt.Println(value2)
			//generate name for the keyed queues
			tempQueueName := ""
			if value2.Keys == nil {
				//		fmt.Print(key)
				//		fmt.Print(selector)
				if key == selector {
					tempQueueName = key
					//fmt.Println(tempQueueName)
					absWorker := Create(tempQueueName)
					//fmt.Print("count")
					//fmt.Print(count)
					outWorkers[count] = absWorker
					//fmt.Println(absWorker)
					count++
				}
			} else {
				for _, keyValue := range value2.Keys {
					if (key + "." + keyValue) == selector {
						tempQueueName = key + "." + keyValue
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
			//fmt.Println(value2)

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

	response = messaging.ServiceResponse{}
	for _, worker := range workerList {
		request.Log("Executing Single Worker : " + worker.GetWorkerName())
		fmt.Println("Executing Single Worker : " + worker.GetWorkerName())
		tmpResponse := ExecuteSingleWorker(request, worker)

		if tmpResponse.IsSuccess {
			request.Log("Executing worker : " + worker.GetWorkerName() + " - Success")
			response.IsSuccess = true
			response.Message = tmpResponse.Message
			response.Stack = tmpResponse.Stack
		} else {
			request.Log("Executing worker : " + worker.GetWorkerName() + " - Failed")
			response.IsSuccess = false
			response.Message = tmpResponse.Message
			response.Stack = tmpResponse.Stack
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
	//conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	host := request.Configuration.ServerConfiguration["DuoV6ServiceServer"]["Host"]
	port := request.Configuration.ServerConfiguration["DuoV6ServiceServer"]["Port"]
	username := request.Configuration.ServerConfiguration["DuoV6ServiceServer"]["UserName"]
	password := request.Configuration.ServerConfiguration["DuoV6ServiceServer"]["Password"]

	conn, err := amqp.Dial("amqp://" + username + ":" + password + "@" + host + ":" + port + "/")
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
					//	fmt.Println(keyValue)
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

	fmt.Print("Running Exchanges : ")
	fmt.Println(myExchanges)
	fmt.Print("Running Queues : ")
	fmt.Println(QueueNameArray)

	listExchanges := ""
	listQueues := ""

	for x := 0; x < len(myExchanges); x++ {
		listExchanges += myExchanges[x]
	}

	for x := 0; x < len(QueueNameArray); x++ {
		listQueues += QueueNameArray[x]
	}

	request.Log("Running Exchanges :")
	request.Log(listExchanges)
	request.Log("Running Queues :")
	request.Log(listQueues)

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
			//fmt.Println("---------------------")
			//log.Printf("ProcessManager : Received a message: %s", d.Body)
			//fmt.Println("----------------------------")
			request.Body = d.Body

			var tempConf = configuration.ConfigurationManager{}.Get()
			var storedServiceConfiguration = configuration.StoreServiceConfiguration{}
			storedServiceConfiguration = tempConf
			request.Configuration = storedServiceConfiguration

			convertedWorkers := getRequiredWorkers(request, d.RoutingKey)
			response := startAtomicOperation(request, convertedWorkers)

			if response.IsSuccess {
				request.Log("Successfully Completed service run!")
				request.Log(response.Message)
				fmt.Println("Successfully Completed service run!")
				fmt.Println(response.Message)

			} else {
				request.Log("Failed service run!")
				request.Log(response.Message)
				fmt.Println("Failed service run!")
				fmt.Println(response.Message)
			}

			if request.MessageStack != nil {
				response.Stack = request.MessageStack
			}

		}

	}()
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Println("Error occured in RabbitMQ! : " + msg)
	}
}
