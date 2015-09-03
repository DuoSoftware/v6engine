package commands

import (
	"time"
)

type Matrices struct {
	TotalMemory     int
	UsedMemory      int
	Freememory      int
	BufferSize      int
	TotalSwapMemory int
	UsedSwapMemory  int
	FeeSwapMemory   int
	SystemTime      time.Time
	SystemupTime    float64
	//systemCpuAvg    systemstat.CPUAverage
}
