package commands

import (
	"time"
)

type Matrices struct {
	TotalMemory     uint64
	UsedMemory      uint64
	Freememory      uint64
	BufferSize      uint64
	TotalSwapMemory uint64
	UsedSwapMemory  uint64
	FeeSwapMemory   uint64
	SystemTime      time.Time
	SystemupTime    float64
	//systemCpuAvg    systemstat.CPUAverage
}
