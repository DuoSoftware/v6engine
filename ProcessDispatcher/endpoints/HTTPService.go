package endpoints

import (
	"duov6.com/objectstore/client"
	//"duov6.com/serviceconsole/common"
	"duov6.com/serviceconsole/messaging"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type HTTPService struct {
}

type ProcessObject struct {
	Id          string
	requestBody messaging.ServiceRequest
}

func (h *HTTPService) Start() {
	fmt.Println("Process Dispatcher Listening on Port : 5000")
	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"securityToken", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	m.Post("/:namespace/:class", handleRequest)

	m.RunOnAddr(":5000")
}

func handleRequest(params martini.Params, res http.ResponseWriter, req *http.Request) { // res and req are injected by Martini

	//start RabbitMQ Pusher

	fmt.Println(params)

	var requestBody1 messaging.ServiceRequest

	rb, _ := ioutil.ReadAll(req.Body)

	err := json.Unmarshal(rb, &requestBody1)

	if err != nil {
		fmt.Println("Error decoding from Json to Struct")
		fmt.Println(err.Error())
	} else {

		publishToRabbitMQ(requestBody1.OperationCode, requestBody1)

		if requestBody1.ScheduleTimeStamp != "" {
			//Push to ObjectStore
			tmp := ProcessObject{}
			temp1 := requestBody1.RefID
			temp2 := requestBody1.RefType
			tmp.Id = (temp1 + temp2)
			tmp.requestBody = requestBody1
			client.Go("token", "schedule", "newobject").StoreObject().WithKeyField("Id").AndStoreOne(tmp).Ok()
		}
	}

}

func publishToRabbitMQ(RoutingKey string, body messaging.ServiceRequest) {

	//get settings

	content, err := ioutil.ReadFile("ProcessDispatcher.config")
	if err != nil {
		//Do something for error
		fmt.Println("FATAL ERROR! ProcessDispatcher.CONFIG file NOT FOUND!")
	} else {
		fmt.Println("Process Dispatcher Configuration loaded successfully!")
	}
	lines := strings.Split(string(content), "\n")

	conn, err := amqp.Dial("amqp://" + lines[0] + ":" + lines[1] + "@" + lines[2] + "/")
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

	body2, _ := json.Marshal(body)

	err = ch.Publish(
		"publisher_01", // exchange
		RoutingKey,     // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body2,
			//Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")

	log.Printf("Data pushed to " + RoutingKey + " Queue")
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Println(msg + " : " + err.Error())
	}
}
