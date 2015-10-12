package statgen

import (
	"strconv"
	"time"
)

func getGenericStatistics() Information {
	retInformation := Information{}

	return retInformation
}

func getTotalCalls(ip string) (retTotalCalls int) {
	//Get all Files with No Filtering
	fileList := getAllFiles()

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retTotalCalls++
			}
		}
	}
	return
}

func getSuccessCalls(ip string) (retSuccessCalls int) {
	//Get success files
	fileList := getSuccessFiles()

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retSuccessCalls++
			}
		}
	}
	return
}

func getFailedCalls(ip string) (retFiledCalls int) {
	//Get failed files
	fileList := getErrorFiles()

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retFiledCalls++
			}
		}
	}
	return
}

func getTotalObjectSize(ip string) (retObjectSize int) {
	//Get all Files with No Filtering
	fileList := getAllFiles()

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retObjectSize = (retObjectSize + singleRecord.ObjectSize)
			}
		}
	}
	return
}

func getElapsedTime(ip string) (retElapsedTime int64) {
	//Get all Files with No Filtering
	fileList := getAllFiles()

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retElapsedTime = (retElapsedTime + singleRecord.ElapsedTime)
			}
		}
	}
	return
}

func getRequestRatio(ip string) (perHour float32, perMinute float32, perSecond float32) {

	thisYear, thisMonth, thisDay := time.Now().Date()

	fileNameSuc := (strconv.Itoa(thisYear) + getMonth(thisMonth) + strconv.Itoa(thisDay) + ".suc")
	fileNameErr := (strconv.Itoa(thisYear) + getMonth(thisMonth) + strconv.Itoa(thisDay) + ".err")

	sucFiles := getFilesByPattern(fileNameSuc)
	errFiles := getFilesByPattern(fileNameErr)

	var fileNames map[int]string
	fileNames = make(map[int]string)

	index := 0

	for _, value := range sucFiles {
		fileNames[index] = value
		index++
	}

	for _, value := range errFiles {
		fileNames[index] = value
		index++
	}

	count := 0
	for _, value := range fileNames {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				count++
			}
		}
	}

	rate := float32(count)

	hour, min, sec := time.Now().Clock()

	hourDivident := hour
	minDivident := ((hour * 60) + min)
	secDivident := (((hour * 60 * 60) + (min * 60)) + sec)

	perHour = rate / float32(hourDivident)
	perMinute = rate / float32(minDivident)
	perSecond = rate / float32(secDivident)

	return
}

func getMonth(month time.Month) (num string) {
	if month.String() == "January" {
		num = "01"
	} else if month.String() == "February" {
		num = "02"
	} else if month.String() == "March" {
		num = "03"
	} else if month.String() == "April" {
		num = "04"
	} else if month.String() == "May" {
		num = "05"
	} else if month.String() == "June" {
		num = "06"
	} else if month.String() == "July" {
		num = "07"
	} else if month.String() == "August" {
		num = "08"
	} else if month.String() == "September" {
		num = "09"
	} else if month.String() == "October" {
		num = "10"
	} else if month.String() == "November" {
		num = "11"
	} else if month.String() == "December" {
		num = "12"
	}

	return
}
