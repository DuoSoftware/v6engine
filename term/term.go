package term

import (
	"duov6.com/updater"
	"fmt"
	"log"
	"os/exec"
	"time"
)

const (
	Reset      = "\x1b[0m"
	Bright     = "\x1b[1m"
	Dim        = "\x1b[2m"
	Underscore = "\x1b[4m"
	Blink      = "\x1b[5m"
	Reverse    = "\x1b[7m"
	Hidden     = "\x1b[8m"

	FgBlack   = "\x1b[30m"
	FgRed     = "\x1b[31m"
	FgGreen   = "\x1b[32m"
	FgYellow  = "\x1b[33m"
	FgBlue    = "\x1b[34m"
	FgMagenta = "\x1b[35m"
	FgCyan    = "\x1b[36m"
	FgWhite   = "\x1b[37m"

	BgBlack   = "\x1b[40m"
	BgRed     = "\x1b[41m"
	BgGreen   = "\x1b[42m"
	BgYellow  = "\x1b[43m"
	BgBlue    = "\x1b[44m"
	BgMagenta = "\x1b[45m"
	BgCyan    = "\x1b[46m"
	BgWhite   = "\x1b[47m"

	Error       = 1
	Information = 0
	Debug       = 2
)

func Read(Lable string) string {
	var S string
	fmt.Printf(FgGreen + Lable + FgMagenta + " LDS$ " + Reset)
	fmt.Scanln(&S)
	//fmt.
	//BgGreen
	return S
}

func Write(Lable string, mType int) {
	//var S string
	switch mType {
	case 1:
		//log.Printf(format, ...)
		fmt.Println(FgRed + time.Now().String() + "Error! " + Lable + Reset)
	case 0:
		fmt.Println(FgGreen + time.Now().String() + "Information! " + Lable + Reset)
	case 2:
		fmt.Println(FgBlue + time.Now().String() + "Debug! " + Lable + Reset)
	default:
		fmt.Println(FgMagenta + time.Now().String() + Lable + Reset)
	}
}

func StartCommandLine() {
	s := Read("Command ")
	for s != "exit" {
		cmd := exec.Command(s, "")
		cmd.Start()
		switch s {
		case "download":
			//Write("Invalid command.", Error)
			updater.DownloadFromUrl(Read("URL"), Read("FileName"))
		default:
			Write("Invalid command.", Error)
		}
		s = Read("Command ")
	}
}
