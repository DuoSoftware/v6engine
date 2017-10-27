package core

import (
	"bytes"
	"duov6.com/serviceconsole/scheduler/common"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"net/http"
	"strconv"
	"strings"
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
	fmt.Println("Executing Dispatcher::AddObjects Method!")
	for _, ob := range objects {
		d.ScheduleTable.InsertObject(ob)
	}
}

func (t *ScheduleTable) Get(timestamp string) (obj []map[string]interface{}) {
	fmt.Println("Executing Dispatcher::Get Object By TimeStamp Method!")
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
	fmt.Println("Executing Dispatcher::Get Row by TimeStamp Method!")
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
	fmt.Println("Executing Dispatcher::InsertObject Method!")
	timestamp := obj["TimeStamp"].(string)

	if t.Contains(timestamp) {
		currentTableRow := t.GetRow(timestamp)
		newObjs := append(currentTableRow.Objects, obj)
		currentTableRow.Objects = newObjs
		t.Delete(timestamp)
		t.AddRow(currentTableRow)
	} else {
		currentTableRow := TableRow{Timestamp: timestamp, Objects: make([]map[string]interface{}, 1)}
		currentTableRow.Objects[0] = obj
		//t.Rows = append(t.Rows, currentTableRow)
		t.AddRow(&currentTableRow)
	}
}

func (t *ScheduleTable) AddRow(row *TableRow) {
	fmt.Println("Executing Dispatcher::AddRow Method!")
	//tablesize := len(t.Rows)
	//t.Rows[tablesize].Timestamp = row.Timestamp
	//t.Rows[tablesize].Objects = row.Objects
	t.Rows = append(t.Rows, *row)
}

func (t *ScheduleTable) Contains(timestamp string) bool {
	fmt.Println("Executing Dispatcher::Contain Method!")
	for _, rows := range t.Rows {
		if rows.Timestamp == timestamp {
			return true
		}
	}
	return false
}

func (t *ScheduleTable) Delete(timestamp string) {
	fmt.Println("Executing Dispatcher::Delete Method!")
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

//Original Method... Don't Delete
// func (t *ScheduleTable) GetForExecution(timestamp string) *TableRow {
// 	fmt.Println("Executing Dispatcher::GetForExecution Method!")
// 	for _, row := range t.Rows {
// 		if row.Timestamp == timestamp {
// 			return &row
// 		}
// 	}

// 	return nil
// }

//working version of earlier. dont delete
// func (t *ScheduleTable) GetForExecution(timestamp string) *TableRow {
// 	fmt.Println("Executing Dispatcher::GetForExecution Method!")

// 	for _, row := range t.Rows {
// 		if strings.Contains(row.Timestamp, timestamp) {
// 			return &row
// 		}

// 		rowTime := t.GetTimeFromString(row.Timestamp)
// 		nowTime := t.GetTimeFromString(timestamp)

// 		if rowTime.Before(nowTime) {
// 			fmt.Println("Before")
// 			return &row
// 		}
// 	}

// 	return nil
// }

func (t *ScheduleTable) GetForExecution(timestamp string) []TableRow {
	fmt.Println("Executing Dispatcher::GetForExecution Method!")

	var tableRowArray []TableRow

	for _, row := range t.Rows {
		if strings.Contains(row.Timestamp, timestamp) {
			tableRowArray = append(tableRowArray, row)
			t.Delete(row.Timestamp)
			//return &row
		}

		rowTime := t.GetTimeFromString(row.Timestamp)
		nowTime := t.GetTimeFromString(timestamp)

		if rowTime.Before(nowTime) {
			fmt.Println("Adding older objects to executing list....")
			//return &row
			tableRowArray = append(tableRowArray, row)
			t.Delete(row.Timestamp)
		}
	}

	return tableRowArray
}

func (t *ScheduleTable) GetTimeFromString(timestamp string) time.Time {

	year, _ := strconv.Atoi(timestamp[0:4])
	month := timestamp[4:6]
	date, _ := strconv.Atoi(timestamp[6:8])
	hour, _ := strconv.Atoi(timestamp[8:10])
	min, _ := strconv.Atoi(timestamp[10:12])

	var monthTime time.Month

	switch month {
	case "01":
		monthTime = time.January
	case "02":
		monthTime = time.February
	case "03":
		monthTime = time.March
	case "04":
		monthTime = time.April
	case "05":
		monthTime = time.May
	case "06":
		monthTime = time.June
	case "07":
		monthTime = time.July
	case "08":
		monthTime = time.August
	case "09":
		monthTime = time.September
	case "10":
		monthTime = time.October
	case "11":
		monthTime = time.November
	case "12":
		monthTime = time.December
	}

	newTime := time.Date(year, monthTime, date, hour, min, 0, 0, time.UTC)

	//newTime, _ := time.Parse("200601021504", timestamp)
	return newTime
}

// func (t *ScheduleTable) GetTimeFromString(timestamp string) *TableRow {

// 	year, _ := strconv.Atoi(timestamp[0:4])
// 	month, _ := strconv.Atoi(timestamp[4:6])
// 	date, _ := strconv.Atoi(timestamp[6:8])
// 	hour, _ := strconv.Atoi(timestamp[8:10])
// 	min, _ := strconv.Atoi(timestamp[10:12])

// 	return
// }

func newDispatcher() (d *Dispatcher) {
	fmt.Println("Executing Dispatcher::NewDispatcher Method!")
	newObj := Dispatcher{}
	newObj.ScheduleTable = ScheduleTable{}
	newObj.ScheduleTable.Rows = make([]TableRow, 0)
	return &newObj
}

func (d *Dispatcher) TriggerTimer() {
	fmt.Println("Executing Dispatcher::TriggerTimer Method!")
	currenttime := time.Now().Local()
	x := currenttime.Format("200601021504")

	tableRows := d.ScheduleTable.GetForExecution(x)

	if len(tableRows) > 0 {
		//dispatchObjectToRabbitMQ(tableRow.Objects)
		for _, tableSingleRow := range tableRows {
			for _, obj := range tableSingleRow.Objects {
				dispatchToTaskQueue(obj)
			}
		}
		//d.ScheduleTable.Delete(tableRow.Timestamp)
	} else {
		fmt.Println("No Objects To Execute at : " + x)
		if len(d.ScheduleTable.Rows) > 0 {
			fmt.Println("But Queued these Tasks : ")
			fmt.Println(d.ScheduleTable.Rows)
		}
	}
}

func dispatchToTaskQueue(object map[string]interface{}) {
	fmt.Println("Executing Dispatcher::Dispatch to Task Queue Method!")
	byteArray, _ := json.Marshal(object)
	settings := common.GetSettings()
	url := settings["SVC_TQ_URL"]
	//url = "http://localhost:6000/aa/bb"
	fmt.Println(url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byteArray))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Data sending error : " + err.Error())
	} else {
		fmt.Println("Data Sent Successfully!")
	}
	defer resp.Body.Close()
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
