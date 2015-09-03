package statservice

import (
	//"duov6.com/common"
	"duov6.com/gorest"
	"duov6.com/stat"
	//"net/http/pprof"
	//"runtime/pprof"
	//"encoding/json"
)

type StatusAll struct {
	NumberOfCalls    int
	TotalSize        int
	TotalElapsedTime int64
}

type StatSvc struct {
	gorest.RestService
	getStatus gorest.EndPoint `method:"GET" path:"stat/GetStatus/{ID:string}" output:"StatusAll"`
}

func (A StatSvc) GetStatus(ID string) StatusAll {

	s := stat.GetStatus(ID)

	return StatusAll{NumberOfCalls: s.NumberOfCalls, TotalElapsedTime: s.TotalElapsedTime, TotalSize: s.TotalSize}
}

func (A StatSvc) GetMemoryUsed() {

}
