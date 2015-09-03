package core

import (
	"duov6.com/objectstore/client"

	"encoding/json"
	"fmt"
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

	nowTime := time.Now().Local()                             //current time
	addedtime := nowTime.Add(time.Duration(15 * time.Minute)) //add 15 minutes to current time
	formattedTime := addedtime.Format("20060102150405")       //formatted new time
	rawBytes, err := client.Go("efba1d6c3566e9bcfdf61a9a8d238dd8", "com.duosoftware.com", "schedule").GetMany().ByQuerying("Timestamp:[* " + formattedTime + "]").Ok()
	if len(err) != 0 {
		fmt.Println("ERROR : " + err)
	}
	d.executeObjects(rawBytes)
}

func (d *Downloader) StartDownloadTimer() { //call downloadObjects every 15 minutes

	c := time.Tick(1 * time.Second /*Minute*/)
	for now := range c {
		_ = now
		d.DownloadTicks++
		d.Dispatcher.TriggerTimer()
		if d.DownloadTicks == 5 {
			d.DownloadTicks = 0
			d.DownloadObjects()
		}
	}
}

func (d *Downloader) executeObjects(raw []byte) {
	unmarshall := make([]map[string]interface{}, 0)
	err := json.Unmarshal(raw, &unmarshall)

	if err != nil {
		fmt.Println("JSON Unmarshll error" + err.Error())
	}
	d.Dispatcher.addObjects(unmarshall)
}
