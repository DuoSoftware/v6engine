package statgen

import (
	"time"
)

type Information struct {
	TotalCalls       int
	SucessCalls      int
	FailedCalls      int
	TotalObjectSize  int
	TotalElapsedTime int64
}

type record struct {
	ID          string
	NameSpace   string
	MethodName  string
	ClientIP    string
	ElapsedTime int64
	CreatedTime time.Time
	Status      int
	ObjectSize  int
}
