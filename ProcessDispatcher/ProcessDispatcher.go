package main

import (
	"duov6.com/ProcessDispatcher/endpoints"
	"duov6.com/fws"
	"fmt"
)

func main() {
	splash()
	initialize()
}

func initialize() {
	fws.Attach("ProcessDispatcher")
	httpServer := endpoints.HTTPService{}
	httpServer.Start()
}

func splash() {
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("______            ______ _                 _       _               ")
	fmt.Println("|  _  \\           |  _  (_)               | |     | |              ")
	fmt.Println("| | | |_   _  ___ | | | |_ ___ _ __   __ _| |_ ___| |__   ___ _ __ ")
	fmt.Println("| | | | | | |/ _ \\| | | | / __| '_ \\ / _` | __/ __| '_ \\ / _ \\ '__|")
	fmt.Println("| |/ /| |_| | (_) | |/ /| \\__ \\ |_) | (_| | || (__| | | |  __/ |   ")
	fmt.Println("|___/  \\__,_|\\___/|___/ |_|___/ .__/ \\__,_|\\__\\___|_| |_|\\___|_|   ")
	fmt.Println("                              | |                                  ")
	fmt.Println("                              |_|           ")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")
}
