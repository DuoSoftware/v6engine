package core

import (
	"duov6.com/objectstore/client"
	"duov6.com/serviceconsole/scheduler/common"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Downloader struct {
	Dispatcher    Dispatcher
	DownloadTicks int
}

func (d *Downloader) Start() {
	d.DownloadObjects()
	d.StartDownloadTimer()
	d.Dispatcher = *newDispatcher()
}

func (d *Downloader) DownloadObjects() {
	fmt.Println("Executing Downloader::DownloadObjects Method!")
	nowTime := time.Now().Local()

	settings := common.GetSettings()
	Timeout, _ := strconv.Atoi(settings["SCHEDULE_CHECK_TIMEOUT"])

	fmt.Println("Checking for objects every " + settings["SCHEDULE_CHECK_TIMEOUT"] + " minutes!")
	//current time
	addedtime := nowTime.Add(time.Duration(time.Duration(Timeout) * time.Minute)) //add 15 minutes to current time
	formattedTime := addedtime.Format("20060102150405")
	fmt.Println(formattedTime) //formatted new time

	namespace := settings["SCHEDULER_NAMESPACE"]
	class := settings["SCHEDULER_CLASS"]

	rawBytes, err := client.Go("efba1d6c3566e9bcfdf61a9a8d238dd8", namespace, class).GetMany().ByQuerying("SELECT * from " + class + " WHERE TimeStamp < " + formattedTime + ";").Ok()

	if len(err) != 0 {
		fmt.Println("ERROR : " + err)
	}

	if len(rawBytes) > 4 {
		fmt.Println("Objects Found for Scheduled Execution : ")
		fmt.Println(string(rawBytes))
		d.DeleteFromObjectStore(rawBytes, namespace, class)
		d.RecurringSchedule(rawBytes, namespace, class)
		d.executeObjects(rawBytes)
	}
}

func (d *Downloader) StartDownloadTimer() { //call downloadObjects every 15 minutes
	fmt.Println("Executing Downloader::Start Download Timer Method!")
	settings := common.GetSettings()
	Timeout, _ := strconv.Atoi(settings["SCHEDULE_CHECK_TIMEOUT"])

	//first time run
	d.Dispatcher.TriggerTimer()

	c := time.Tick(time.Duration(Timeout/Timeout) * time.Minute /*Minute*/)
	for now := range c {
		_ = now
		d.DownloadTicks++
		d.Dispatcher.TriggerTimer()
		if d.DownloadTicks == Timeout {
			fmt.Println("End of Time Out! Redownload!")
			d.DownloadTicks = 0
			d.DownloadObjects()
		}
	}
}

//Original Method... Don't Delete
// func (d *Downloader) StartDownloadTimer() { //call downloadObjects every 15 minutes
// 	fmt.Println("Executing Downloader::Start Download Timer Method!")
// 	//settings := common.GetSettings()
// 	//Timeout, _ := strconv.Atoi(settings["SCHEDULE_CHECK_TIMEOUT"])

// 	c := time.Tick(1 * time.Minute /*Minute*/)
// 	for now := range c {
// 		fmt.Println("PING")
// 		_ = now
// 		d.DownloadTicks++
// 		d.Dispatcher.TriggerTimer()
// 		if d.DownloadTicks == 5 {
// 			fmt.Println("End of Time Out! Re-download!")
// 			d.DownloadTicks = 0
// 			d.DownloadObjects()
// 		}
// 	}
// }

func (d *Downloader) executeObjects(raw []byte) {
	fmt.Println("Executing Downloader::executeObjects Method!")
	var unmarshall []map[string]interface{}
	err := json.Unmarshal(raw, &unmarshall)

	if err != nil {
		fmt.Println("JSON Unmarshll error : " + err.Error())
	} else {
		d.Dispatcher.addObjects(unmarshall)
	}
}

func (d *Downloader) DeleteFromObjectStore(raw []byte, namespace string, class string) {
	fmt.Println("Executing Downloader::DeleteFromObjectStore Method!")
	unmarshall := make([]map[string]interface{}, 0)
	err := json.Unmarshal(raw, &unmarshall)

	if err != nil {
		fmt.Println("JSON Unmarshll error : " + err.Error())
	} else {
		for _, obj := range unmarshall {
			fmt.Println("Deleting fetched objects from Objects!")
			client.Go("token", namespace, class).DeleteObject().WithKeyField("RefId").AndDeleteOne(obj).Ok()
		}
	}
}

func (d *Downloader) RecurringSchedule(raw []byte, namespace string, class string) {
	fmt.Println("Executing Downloader::Recurring Schedule Method!")
	unmarshall := make([]map[string]interface{}, 0)
	err := json.Unmarshal(raw, &unmarshall)

	var saveList []map[string]interface{}

	if err != nil {
		fmt.Println("JSON Unmarshll error : " + err.Error())
	} else {
		for _, obj := range unmarshall {
			timeStamp, occuranceCount := d.CheckIfRecurring(obj)
			if timeStamp != "" {
				obj["TimeStamp"] = timeStamp
				obj["TimeStampReadable"] = d.GetReadableTimeStamp(timeStamp)
				ScheduleParameters := obj["ScheduleParameters"].(map[string]interface{})
				ScheduleParameters["OccuranceCount"] = occuranceCount
				obj["ScheduleParameters"] = ScheduleParameters
				saveList = append(saveList, obj)
			}
		}

		if len(saveList) > 0 {
			client.Go("ignore", namespace, class).StoreObject().WithKeyField("RefId").AndStoreManyObjects(saveList).Ok()
		}
	}
}

func (d *Downloader) CheckIfRecurring(obj map[string]interface{}) (timestamp string, occurunceCount int) {
	ScheduleParameters := make(map[string]interface{})
	ScheduleParameters = obj["ScheduleParameters"].(map[string]interface{})
	ScheduleQty := 1
	ScheduleType := ""
	timestamp = ""
	objectTimeStamp := obj["TimeStamp"].(string)
	occurunceCount = int(ScheduleParameters["OccuranceCount"].(float64))

	if ScheduleParameters["ScheduleQty"] != nil {
		if ScheduleParameters["OccuranceCount"].(float64) > 1 && ScheduleParameters["ScheduleQty"].(float64) > 1 {
			ScheduleQty = int(ScheduleParameters["ScheduleQty"].(float64))
			occurunceCount -= 1
		} else {
			timestamp = ""
			return
		}
	}

	if ScheduleParameters["ScheduleType"] != nil {
		ScheduleType = ScheduleParameters["ScheduleType"].(string)
	}

	timeInTime := d.GetTimeFromString(objectTimeStamp)

	keyword := strings.ToLower(ScheduleType)

	switch keyword {
	case "hourly":
		newTime := timeInTime.Add(time.Duration(ScheduleQty) * time.Hour)
		timestamp = newTime.Format("20060102150405")
		break
	case "daily":
		newTime := timeInTime.Add((time.Duration(24 * ScheduleQty)) * time.Hour)
		timestamp = newTime.Format("20060102150405")
		break
	case "weekly":
		newTime := timeInTime.Add((time.Duration(24 * 7 * ScheduleQty)) * time.Hour)
		timestamp = newTime.Format("20060102150405")
		break
	case "monthly":
		newTime := timeInTime.Add(time.Duration(24*30*ScheduleQty) * time.Hour)
		timestamp = newTime.Format("20060102150405")
		break
	case "yearly":
		newTime := timeInTime.Add(time.Duration(24*30*12*ScheduleQty) * time.Hour)
		timestamp = newTime.Format("20060102150405")
		break
	default:
		timestamp = ""
		break
	}

	return
}

func (d *Downloader) GetReadableTimeStamp(timestamp string) (readableTimeStamp string) {
	readableTimeStamp = timestamp[0:4] + "-" + timestamp[4:6] + "-" + timestamp[6:8] + " " + timestamp[8:10] + ":" + timestamp[10:12] + ":" + timestamp[12:14]
	return
}

func (d *Downloader) GetTimeFromString(timestamp string) time.Time {

	year, _ := strconv.Atoi(timestamp[0:4])
	month := timestamp[4:6]
	date, _ := strconv.Atoi(timestamp[6:8])
	hour, _ := strconv.Atoi(timestamp[8:10])
	min, _ := strconv.Atoi(timestamp[10:12])
	seconds, _ := strconv.Atoi(timestamp[12:14])

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

	newTime := time.Date(year, monthTime, date, hour, min, seconds, 0, time.UTC)
	return newTime
}
