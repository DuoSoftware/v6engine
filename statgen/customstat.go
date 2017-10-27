package statgen

import (
	"strconv"
)

func getTotalHitsByDay(ip string, year int, month int, day int) (retHits int) {

	sYear := strconv.Itoa(year)
	sMonth := strconv.Itoa(month)
	sDay := strconv.Itoa(day)

	var fileList map[int]string
	fileList = make(map[int]string)

	fileListErr := getFilesByPattern((sYear + sMonth + sDay + ".err"))
	fileListSuc := getFilesByPattern((sYear + sMonth + sDay + ".suc"))

	index := 0

	for _, value := range fileListErr {
		fileList[index] = value
		index++
	}

	for _, value := range fileListSuc {
		fileList[index] = value
		index++
	}

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retHits++
			}
		}
	}
	return
}

func getTotalHitsByMonth(ip string, year int, month int) (retHits int) {

	sYear := strconv.Itoa(year)
	sMonth := strconv.Itoa(month)

	var fileList map[int]string
	fileList = make(map[int]string)

	fileListErr := getFilesByPattern((sYear + sMonth + "*.err"))
	fileListSuc := getFilesByPattern((sYear + sMonth + "*.suc"))

	index := 0

	for _, value := range fileListErr {
		fileList[index] = value
		index++
	}

	for _, value := range fileListSuc {
		fileList[index] = value
		index++
	}

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retHits++
			}
		}
	}
	return
}

func getTotalHitsByYear(ip string, year int) (retHits int) {
	sYear := strconv.Itoa(year)

	var fileList map[int]string
	fileList = make(map[int]string)

	fileListErr := getFilesByPattern((sYear + "*.err"))
	fileListSuc := getFilesByPattern((sYear + "*.suc"))

	index := 0

	for _, value := range fileListErr {
		fileList[index] = value
		index++
	}

	for _, value := range fileListSuc {
		fileList[index] = value
		index++
	}

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retHits++
			}
		}
	}
	return
}

// Successive Hits by  Day / Month / Year

func getSuccessHitsByDay(ip string, year int, month int, day int) (retHits int) {

	sYear := strconv.Itoa(year)
	sMonth := strconv.Itoa(month)
	sDay := strconv.Itoa(day)

	fileList := getFilesByPattern((sYear + sMonth + sDay + ".suc"))

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retHits++
			}
		}
	}
	return
}

func getSuccessHitsByMonth(ip string, year int, month int) (retHits int) {

	sYear := strconv.Itoa(year)
	sMonth := strconv.Itoa(month)

	fileList := getFilesByPattern((sYear + sMonth + "*.suc"))

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retHits++
			}
		}
	}
	return
}

func getSuccessHitsByYear(ip string, year int) (retHits int) {

	sYear := strconv.Itoa(year)

	fileList := getFilesByPattern((sYear + "*.suc"))

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retHits++
			}
		}
	}
	return
}

// Failed Hits by  Day / Month / Year

func getFailedHitsByDay(ip string, year int, month int, day int) (retHits int) {

	sYear := strconv.Itoa(year)
	sMonth := strconv.Itoa(month)
	sDay := strconv.Itoa(day)

	fileList := getFilesByPattern((sYear + sMonth + sDay + ".err"))

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retHits++
			}
		}
	}
	return
}

func getFailedHitsByMonth(ip string, year int, month int) (retHits int) {

	sYear := strconv.Itoa(year)
	sMonth := strconv.Itoa(month)

	fileList := getFilesByPattern((sYear + sMonth + "*.err"))

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retHits++
			}
		}
	}
	return
}

func getFailedHitsByYear(ip string, year int) (retHits int) {

	sYear := strconv.Itoa(year)

	fileList := getFilesByPattern((sYear + "*.err"))

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retHits++
			}
		}
	}
	return
}

// Object Data transferres by  Day / Month / Year

func getObjectSizeByDay(ip string, year int, month int, day int) (retHits int) {

	sYear := strconv.Itoa(year)
	sMonth := strconv.Itoa(month)
	sDay := strconv.Itoa(day)

	var fileList map[int]string
	fileList = make(map[int]string)

	fileListErr := getFilesByPattern((sYear + sMonth + sDay + ".err"))
	fileListSuc := getFilesByPattern((sYear + sMonth + sDay + ".suc"))

	index := 0

	for _, value := range fileListErr {
		fileList[index] = value
		index++
	}

	for _, value := range fileListSuc {
		fileList[index] = value
		index++
	}

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retHits = (retHits + singleRecord.ObjectSize)
			}
		}
	}
	return
}

func getObjectSizeByMonth(ip string, year int, month int) (retHits int) {

	sYear := strconv.Itoa(year)
	sMonth := strconv.Itoa(month)

	var fileList map[int]string
	fileList = make(map[int]string)

	fileListErr := getFilesByPattern((sYear + sMonth + "*.err"))
	fileListSuc := getFilesByPattern((sYear + sMonth + "*.suc"))

	index := 0

	for _, value := range fileListErr {
		fileList[index] = value
		index++
	}

	for _, value := range fileListSuc {
		fileList[index] = value
		index++
	}

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retHits = (retHits + singleRecord.ObjectSize)
			}
		}
	}
	return
}

func getObjectSizeByYear(ip string, year int) (retHits int) {

	sYear := strconv.Itoa(year)

	var fileList map[int]string
	fileList = make(map[int]string)

	fileListErr := getFilesByPattern((sYear + "*.err"))
	fileListSuc := getFilesByPattern((sYear + "*.suc"))

	index := 0

	for _, value := range fileListErr {
		fileList[index] = value
		index++
	}

	for _, value := range fileListSuc {
		fileList[index] = value
		index++
	}

	for _, value := range fileList {
		fileData := readFileContent(value)

		//Iterate through data find IP address if match increase count
		for _, singleRecord := range fileData {
			if singleRecord.ClientIP == ip {
				retHits = (retHits + singleRecord.ObjectSize)
			}
		}
	}
	return
}
