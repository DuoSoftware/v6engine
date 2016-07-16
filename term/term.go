package term

import (
	"bufio"
	"duov6.com/config"
	"duov6.com/updater"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	//"time"
	"github.com/fatih/color"
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
	Splash      = 3
	Blank       = 4
)

var Config TerminalConfig
var currentPlugin TermPlugin

func GetConfig() TerminalConfig {
	b, err := config.Get("Terminal")
	if err == nil {
		json.Unmarshal(b, &Config)
	} else {
		Config = TerminalConfig{}
		Config.DebugLine = true
		Config.ErrorLine = true
		Config.InformationLine = true

		config.Add(Config, "Terminal")
	}
	return Config
}

func SetConfig(c TerminalConfig) {
	Config = c
	config.Add(c, "Terminal")
}

func SetupConfig() {

	Config = GetConfig()

	//SplashScreen("setup.art")
	if Read("Do want to Debug (y/n)") == "y" {
		Config.DebugLine = true
	} else {

		Config.DebugLine = false
	}

	if Read("Do want show Errors (y/n)") == "y" {
		Config.ErrorLine = true
	} else {
		Config.ErrorLine = false
	}

	if Read("Do want show Information (y/n)") == "y" {
		Config.InformationLine = true
	} else {
		Config.InformationLine = false
	}
	SetConfig(Config)

}

func Read(Lable string) string {
	var S string
	fmt.Printf(FgGreen + Lable + FgMagenta + " LDS$ " + Reset)
	fmt.Scanln(&S)
	//fmt.
	//BgGreen
	return S
}

func Write(data interface{}, mType int) {

	//fmt.Println(data)

	Lable := ""
	if reflect.TypeOf(data).String() == "string" {
		Lable = data.(string)
	} else {
		byteArray, err := json.Marshal(data)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		Lable = string(byteArray)
	}

	switch mType {
	case Error:
		// if Config.ErrorLine {
		// 	fmt.Println(time.Now().Format("2006-01-02 15:04:05") + FgRed + BgWhite + " Error! " + Reset + Lable + Reset)
		// }
		color.Red(Lable)
	case Information:
		// if Config.InformationLine {
		// 	fmt.Println(FgGreen + time.Now().Format("2006-01-02 15:04:05") + " Information! " + Lable + Reset)
		// }
		color.Yellow(Lable)
	case Debug:
		// if Config.DebugLine {
		// 	fmt.Println(FgBlue + time.Now().Format("2006-01-02 15:04:05") + " Debug! " + Lable + Reset)
		// }
		//fmt.Println(FgBlue + time.Now().Format("2006-01-02 15:04:05") + " Debug!")
		color.Green(Lable)
	case Splash:
		//fmt.Println(FgBlack + BgWhite + Lable + Reset)
	case Blank:
		//fmt.Println(Lable)
		fmt.Println(Lable)
	default:
		//fmt.Println(FgMagenta + time.Now().String() + Lable + Reset)
		fmt.Println(Lable)
	}

	if currentPlugin != nil {
		currentPlugin.Log(Lable, mType)
	}
}

//ORIGINAL WRITE FUNCTION WITH NO SUPPORT FOR STDOUTing STRUCTURES OTHER THAN STRINGS. DONT DELETE!
// func Write(Lable string, mType int) {
// 	//var S string
// 	switch mType {
// 	case Error:
// 		//log.Printf(format, ...)
// 		if Config.ErrorLine {
// 			fmt.Println(time.Now().String() + FgRed + BgWhite + " Error! " + Reset + Lable + Reset)
// 		}
// 	case Information:
// 		if Config.InformationLine {
// 			fmt.Println(FgGreen + time.Now().String() + " Information! " + Lable + Reset)
// 		}
// 	case Debug:
// 		//if Config.DebugLine {
// 		fmt.Println(FgBlue + time.Now().String() + " Debug! " + Lable + Reset)
// 		//}
// 	case Splash:
// 		fmt.Println(FgBlack + BgWhite + Lable + Reset)
// 	case Blank:
// 		fmt.Println(Lable)
// 	default:
// 		fmt.Println(FgMagenta + time.Now().String() + Lable + Reset)
// 	}

// 	if currentPlugin != nil {
// 		currentPlugin.Log(Lable, mType)
// 	}
// }

func SplashScreen(fileName string) {

	file, _ := os.Open(fileName)
	if file != nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			//split key and value
			fmt.Println(FgBlack + BgWhite + scanner.Text() + Reset)
		}
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
		case "config":
			SetupConfig()
		default:
			Write("Invalid command.", Error)
		}
		s = Read("Command ")
	}
}

func AddPlugin(t TermPlugin) {
	currentPlugin = t
}

func RemovePlugin(t TermPlugin) {
	currentPlugin = nil
}

type TerminalConfig struct {
	DebugLine       bool
	ErrorLine       bool
	InformationLine bool
}
