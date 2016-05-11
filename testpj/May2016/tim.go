package main

import (
	"fmt"
	"time"
)

func main() {
	t1 := getTimeFromString()
	t2 := add3minutesToTime()
	duration1 := CheckHowMuchTimeElapsedInMinutes(t1, t2)

	fmt.Println(t1)
	fmt.Println(t2)
	fmt.Println(duration1)
}

func getTimeFromString() (timeString string) {
	nowTime := time.Now().UTC()
	timeString = nowTime.Format("2006-01-02 15:04:05")
	return
}

func add3minutesToTime() (timeString string) {
	nowTime := time.Now().UTC()
	nowTime = nowTime.Add(3 * time.Minute)
	timeString = nowTime.Format("2006-01-02 15:04:05")
	return
}

func CheckHowMuchTimeElapsedInMinutes(time1 string, time2 string) (minutesTime float64) {
	Ttime1, _ := time.Parse("2006-01-02 15:04:05", time1)
	Ttime2, _ := time.Parse("2006-01-02 15:04:05", time2)

	difference := Ttime1.Sub(Ttime2)
	minutesTime = difference.Minutes()
	return
}
