package stat

import (
	"duov6.com/common"
	"encoding/json"
	//"fmt"
	"time"
)

const (
	Error  = -1
	Sucess = 0
	New    = 1
)

var Exit bool
var isRuning bool
var Data []States

type States struct {
	ID          string
	NameSpace   string
	MethodName  string
	ClientIP    string
	ElapsedTime int64
	CreatedTime time.Time
	Status      int
	ObjectSize  int
}

type StatusAll struct {
	NumberOfCalls    int
	TotalSize        int
	TotalElapsedTime int64
}

var SucessFullCall StatusAll
var FailedCall StatusAll

func Add(stats States) string {
	if isRuning {
		if stats.ID == "" {
			stats.ID = common.GetGUID()
			stats.CreatedTime = time.Now()
		}
		Data = append(Data, stats)
	}
	return stats.ID
}

func Start() {
	SucessFullCall = StatusAll{}
	FailedCall = StatusAll{}
	go startProcess()
}

func Stop() {
	Exit = true
}

func GetStatus(ID string) StatusAll {
	if ID == "Error" {
		return FailedCall
	} else {
		return SucessFullCall
	}
}

func startProcess() {
	if !isRuning {
		isRuning = true
		for {
			s := Data
			Data = []States{}
			if len(s) != 0 {
				Succfilename := time.Now().Format("20060102") + ".suc"
				Errfilename := time.Now().Format("20060102") + ".err"

				errorlog := ""
				Sucesslog := ""
				for _, element := range s {
					if element.Status == Error {
						FailedCall.NumberOfCalls++
						FailedCall.TotalElapsedTime += element.ElapsedTime
						FailedCall.TotalSize += element.ObjectSize
						dataset, _ := json.Marshal(element)
						errorlog += string(dataset) + "\n"
					} else {
						SucessFullCall.NumberOfCalls++
						SucessFullCall.TotalElapsedTime += element.ElapsedTime
						SucessFullCall.TotalSize += element.ObjectSize
						dataset, _ := json.Marshal(element)
						Sucesslog += string(dataset) + "\n"
					}
				}
				//fmt.Println(errorlog)
				//fmt.Println(Sucesslog)
				common.SaveFile(Errfilename, errorlog)
				common.SaveFile(Succfilename, Sucesslog)
			}

			if Exit {
				Exit = false
				isRuning = false
				return
			}
			time.Sleep(10 * time.Second)

		}
	}
}
