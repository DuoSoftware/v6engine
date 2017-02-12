package term

import (
	"bufio"
	"duov6.com/config"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"os"
	"reflect"
	"time"
)

var Config TerminalConfig
var currentPlugin TermPlugin

func GetConfig() TerminalConfig {
	//Initialize Config
	b, err := config.Get("Terminal")
	if err == nil {
		json.Unmarshal(b, &Config)
	} else {
		Config = TerminalConfig{}
		Config.DebugLine = false
		Config.ErrorLine = true
		Config.InformationLine = false
		config.Add(Config, "Terminal")
	}

	return Config
}

func SetConfig(c TerminalConfig) {
	Config = c
	config.Add(c, "Terminal")
}

func ToggleConfig() (status string) {
	if !Config.DebugLine && !Config.InformationLine {
		Config.InformationLine = true
		status = "Enabled Information and Warning Logs."
	} else if !Config.DebugLine && Config.InformationLine {
		Config.DebugLine = true
		status = "Enabled Information, Warning and Debug Logs."
	} else if Config.DebugLine && Config.InformationLine {
		Config.InformationLine = false
		Config.DebugLine = false
		status = "Disabled All Logs other than Error Logs."
	}
	SetConfig(Config)
	return
}

func Read(Lable string) string {
	var S string
	fmt.Printf(FgGreen + Lable + FgMagenta + " LDS$ " + Reset)
	fmt.Scanln(&S)
	return S
}

func Write(data interface{}, mType int) {

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
		if Config.ErrorLine {
			color.Red(time.Now().Format("2006-01-02 15:04:05") + " : " + Lable)
		}
	case Information:
		if Config.InformationLine {
			color.Cyan(time.Now().Format("2006-01-02 15:04:05") + " : " + Lable)
		}
	case Debug:
		if Config.DebugLine {
			color.Green(time.Now().Format("2006-01-02 15:04:05") + " : " + Lable)
		}
	case Splash:
		fmt.Println(FgBlack + BgWhite + Lable + Reset)
	case Blank:
		if Config.InformationLine {
			color.Magenta(time.Now().Format("2006-01-02 15:04:05") + " : " + Lable)
		}
	case Warning:
		if Config.InformationLine {
			color.Yellow(time.Now().Format("2006-01-02 15:04:05") + " : " + Lable)
		}
	default:
		color.Blue(time.Now().Format("2006-01-02 15:04:05") + " : " + Lable)
	}

	// if currentPlugin != nil {
	// 	currentPlugin.Log(Lable, mType)
	// }
}

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
	Warning     = 5
)
