package core

import (
	"duov6.com/objectstore/client"
	"duov6.com/serviceconsole/scheduler/common"
	"encoding/json"
	"fmt"
	"strconv"
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
