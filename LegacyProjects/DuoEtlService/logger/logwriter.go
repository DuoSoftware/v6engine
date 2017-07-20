package logger

import (
	"duov6.com/common"
	"fmt"
)

func Log(message string) {
	fmt.Println(message)
	common.PublishLog("DuoETLServiceLog.log", message)
}
