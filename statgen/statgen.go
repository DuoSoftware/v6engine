package statgen

//Generic GET methods go here

func GetGenericStatistics(ip string) Information {
	retInformation := Information{}
	retInformation.TotalCalls = GetTotalCalls(ip)
	retInformation.SucessCalls = GetSuccessCalls(ip)
	retInformation.FailedCalls = GetFailedCalls(ip)
	retInformation.TotalObjectSize = GetTotalObjectSize(ip)
	retInformation.TotalElapsedTime = GetElapsedTime(ip)
	return retInformation
}

func GetTotalCalls(ip string) (retTotalCalls int) {
	retTotalCalls = getTotalCalls(ip)
	return
}

func GetSuccessCalls(ip string) (retSuccessCalls int) {
	retSuccessCalls = getSuccessCalls(ip)
	return
}

func GetFailedCalls(ip string) (retFailedCalls int) {
	retFailedCalls = getFailedCalls(ip)
	return
}

func GetTotalObjectSize(ip string) (retObjectSize int) {
	retObjectSize = getTotalObjectSize(ip)
	return
}

func GetElapsedTime(ip string) (retElapsedTime int64) {
	retElapsedTime = getElapsedTime(ip)
	return
}

func GetRequestsPerHour(ip string) (rate float32) {
	rate, _, _ = getRequestRatio(ip)
	return
}

func GetRequestsPerMinute(ip string) (rate float32) {
	_, rate, _ = getRequestRatio(ip)
	return
}

func GetRequestsPerSecond(ip string) (rate float32) {
	_, _, rate = getRequestRatio(ip)
	return
}

// Get Total Hits by Day / Month / Year

func GetTotalHitsByDay(ip string, year int, month int, day int) (retHits int) {
	retHits = getTotalHitsByDay(ip, year, month, day)
	return
}

func GetTotalHitsByMonth(ip string, year int, month int) (retHits int) {
	retHits = getTotalHitsByMonth(ip, year, month)
	return
}

func GetTotalHitsByYear(ip string, year int) (retHits int) {
	retHits = getTotalHitsByYear(ip, year)
	return
}

// Successive Hits by  Day / Month / Year

func GetSuccessHitsByDay(ip string, year int, month int, day int) (retHits int) {
	retHits = getSuccessHitsByDay(ip, year, month, day)
	return
}

func GetSuccessHitsByMonth(ip string, year int, month int) (retHits int) {
	retHits = getSuccessHitsByMonth(ip, year, month)
	return
}

func GetSuccessHitsByYear(ip string, year int) (retHits int) {
	retHits = getSuccessHitsByYear(ip, year)
	return
}

// Failed Hits by  Day / Month / Year

func GetFailedHitsByDay(ip string, year int, month int, day int) (retHits int) {
	retHits = getFailedHitsByDay(ip, year, month, day)
	return
}

func GetFailedHitsByMonth(ip string, year int, month int) (retHits int) {
	retHits = getFailedHitsByMonth(ip, year, month)
	return
}

func GetFailedHitsByYear(ip string, year int) (retHits int) {
	retHits = getFailedHitsByYear(ip, year)
	return
}

// Object Data transferres by  Day / Month / Year

func GetObjectSizeByDay(ip string, year int, month int, day int) (retHits int) {
	retHits = getObjectSizeByDay(ip, year, month, day)
	return
}

func GetObjectSizeByMonth(ip string, year int, month int) (retHits int) {
	retHits = getObjectSizeByMonth(ip, year, month)
	return
}

func GetObjectSizeByYear(ip string, year int) (retHits int) {
	retHits = getObjectSizeByYear(ip, year)
	return
}
