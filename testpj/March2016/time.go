package main

import (
	"fmt"
	"strconv"
	"time"
)

func main() {

	dd := getTime()
	fmt.Println(dd)

	fmt.Println("--------------------")

	fmt.Println(GetTimeFromString(dd))

}

func getTime() (retTime string) {
	currenttime := time.Now().Local()
	retTime = currenttime.Format("200601021504")

	return
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

	t := time.Date(year, monthTime, date, hour, min, 0, 0, time.UTC)
	t2 := time.Date(year, monthTime, date, hour, 35, 0, 0, time.UTC)

	if t2.Before(t) {
		fmt.Println("YAY")
	} else {
		fmt.Println("NAH")
	}
	return t
}
