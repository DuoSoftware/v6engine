package stat

import (
//"duov6.com/common"
)

type Statistic struct {
	ErrorCount      int
	SuessCount      int
	ErrorDataSize   int
	ErrorTakenTime  int
	SucessDataSize  int
	SucessTakenTime int
}

type ReportProcessor struct {
	//clinetCalls
}

func (R *ReportProcessor) Add(objects States) {
	//objects.
}
