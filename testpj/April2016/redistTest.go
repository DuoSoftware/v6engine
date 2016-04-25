package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xuyu/goredis"
	"strconv"
	"time"
)

func store() {
	client, _ := GetConnection()
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	_ = client.Set("cc", now, 0, 0, false, false)
}

func get() {
	client, _ := GetConnection()
	val, _ := client.Get("ff")
	fmt.Println(val)
	if val == nil {
		fmt.Println("ooooooooo")
	}
	// timeer, err := time.Parse("2006-01-02 15:04:05", string(val))
	// if err != nil {
	// 	fmt.Println(err.Error())
	// } else {
	// 	fmt.Println(timeer)
	// 	fmt.Println(time.Now().UTC())

	// 	diff := time.Now().UTC().Sub(timeer)
	// 	fmt.Println(diff.Minutes())
	// 	if diff.Minutes() > 3 {
	// 		fmt.Println("oh yeah!")
	// 	}
	// }

}

func GetConnection() (client *goredis.Redis, err error) {
	client, err = goredis.DialURL("tcp://@" + "localhost" + ":" + "6379" + "/0?timeout=1s&maxidle=1")
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("Connection to REDIS Failed!")
	}
	return
}

func main() {
	//store()
	delete()
	//get()
	//incr()
	//fmt.Println(CheckKeyGenLock())
}

func incr() {
	client, _ := GetConnection()
	val, _ := client.Incr("cc")

	dd := strconv.FormatInt(val, 16)

	fmt.Println(dd)
}

func delete() {
	client, _ := GetConnection()
	stat, _ := client.Expire("cc", 0)
	fmt.Println(stat)
}

func GetTimeFromString(timestamp string) time.Time {

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
	return newTime
}

func CheckKeyGenLock() (status bool) {
	client, _ := GetConnection()
	status = false
	key := "asdf"
	val, err := client.Get(key)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if val == nil {
		_ = client.Set(key, "true", 0, 0, false, false)
		return
	}

	err = json.Unmarshal(val, &status)
	if err != nil {
		fmt.Println(err.Error())
	}

	return

}
