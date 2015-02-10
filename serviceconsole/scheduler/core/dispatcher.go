package core

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

type Dispatcher struct {
	ScheduleTable ScheduleTable
}
type ScheduleTable struct {
	Rows []TableRow
}
type TableRow struct {
	Timestamp string
	Objects   []map[string]interface{}
}

func (d *Dispatcher) addObjects(objects []map[string]interface{}) {
	for _, ob := range objects {
		d.ScheduleTable.InsertObject(ob)
	}
}

func (t *ScheduleTable) Get(timestamp string) (obj []map[string]interface{}) {
	if t.Contains(timestamp) == true {

		for _, element := range t.Rows {
			if element.Timestamp == timestamp {
				return element.Objects
			}
		}
	}
	return nil
}

func (t *ScheduleTable) GetRow(timestamp string) *TableRow {
	if t.Contains(timestamp) == true {

		for _, element := range t.Rows {
			if element.Timestamp == timestamp {
				return &element
			}
		}
	}

	return nil
}

func (t *ScheduleTable) InsertObject(obj map[string]interface{}) {

	timestamp := obj["Timestamp"].(string)

	if t.Contains(timestamp) {
		currentTableRow := t.GetRow(timestamp)
		newObjs := append(currentTableRow.Objects, obj)
		currentTableRow.Objects = newObjs
	} else {

		currentTableRow := TableRow{Timestamp: timestamp, Objects: make([]map[string]interface{}, 1)}
		currentTableRow.Objects[0] = obj
		//t.Rows = append(t.Rows, currentTableRow)
		t.AddRow(currentTableRow)
	}
}

func (t *ScheduleTable) AddRow(row *TableRow) {
	//tablesize := len(t.Rows)
	//t.Rows[tablesize].Timestamp = row.Timestamp
	//t.Rows[tablesize].Objects = row.Objects
	t.Rows = append(t.Rows, row)
}

func (t *ScheduleTable) Contains(timestamp string) bool {
	for _, rows := range t.Rows {
		if rows.Timestamp == timestamp {
			return true
		}
	}
	return false

}

func (t *ScheduleTable) Delete(timestamp string) {
	var removeIndex = -1

	for index, e := range t.Rows {
		if e.Timestamp == timestamp {
			removeIndex = index
		}
	}

	if removeIndex != -1 {
		t.Rows = append(t.Rows[:removeIndex], t.Rows[removeIndex+1:]...)
	}

}

func (t *ScheduleTable) GetForExecution(timestamp string) *TableRow {
	for _, row := range t.Rows {
		if row.Timestamp == timestamp {
			return &row
		}
	}

	return nil
}

func newDispatcher() (d *Dispatcher) {
	newObj := Dispatcher{}
	newObj.ScheduleTable = ScheduleTable{}
	newObj.ScheduleTable.Rows = make([]TableRow, 0)
	return &newObj

}

func (d *Dispatcher) TriggerTimer() {
	currenttime := time.Now().Local()
	x := currenttime.Format("20141212101112")
	tableRow := d.ScheduleTable.GetForExecution(x)
	if tableRow != nil {
		dispatchObjectToRabbitMQ(tableRow.Objects)
		d.ScheduleTable.Delete(tableRow.Timestamp)
	}
}

func dispatchObjectToRabbitMQ(objects []map[string]interface{}) {
	fmt.Println("dispatchtorabbitmq method")
	conn, err := amqp.Dial("amqp://admin:admin@192.168.1.194:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	ch.ExchangeDeclare("v6Exchange", "direct", true, false, false, false, nil)
	q, err := ch.QueueDeclare(
		"DuoRabbitMq", // name
		true,          // durable
		false,         // delete when usused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	ch.QueueBind("DuoRabbitMq", "DuoRabbitMq", "v6Exchange", false, nil)

	failOnError(err, "Failed to declare a queue")

	for _, transfer := range objects {
		dataset, _ := json.Marshal(transfer)
		body := dataset
		err = ch.Publish(
			"v6Exchange", // exchange
			q.Name,       // routing key
			false,        // mandatory
			false,        // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})

	}

	failOnError(err, "Failed to publish a message")

}

func failOnError(err error, msg string) {
	if err != nil {
		//log.Fatalf("%s: %s", msg, err)
		fmt.Println(err.Error())
		//panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
