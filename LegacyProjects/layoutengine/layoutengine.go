package main

import (
	"duov6.com/layoutengine/endpoint"
	//"duov6.com/fws"
	"fmt"
)

func main() {
	splash()
	initialize()
}

func initialize() {
	//fws.Attach("layoutengine")
	httpServer := endpoint.HTTPService{}
	httpServer.Start()
}

func splash() {
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")
	fmt.Println(" ______             _                             _   ")
	fmt.Println("|  _  \\           | |                           | |  ")
	fmt.Println("| | | |_   _  ___ | |     __ _ _   _  ___  _   _| |_ ")
	fmt.Println("| | | | | | |/ _ \\| |    / _` | | | |/ _ \\| | | | __|")
	fmt.Println("| |/ /| |_| | (_) | |___| (_| | |_| | (_) | |_| | |_ ")
	fmt.Println("|___/  \\__,_|\\___/\\_____/\\__,_|\\__, |\\___/ \\__,_|\\__|")
	fmt.Println("                                __/ |                ")
	fmt.Println("                               |___/ ")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")
}
