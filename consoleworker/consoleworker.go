package main

import (
	"duov6.com/consoleworker/endpoints"
	"fmt"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	Splash()

	httpServer := endpoints.HTTPService{}
	httpServer.Start()
}

func Splash() {
	fmt.Println()
	fmt.Println()
	fmt.Println("______             _    _            _             ")
	fmt.Println("|  _  \\           | |  | |          | |            ")
	fmt.Println("| | | |_   _  ___ | |  | | ___  _ __| | _____ _ __ ")
	fmt.Println("| | | | | | |/ _ \\| |/\\| |/ _ \\| '__| |/ / _ \\ '__|")
	fmt.Println("| |/ /| |_| | (_) \\  /\\  / (_) | |  |   <  __/ |   ")
	fmt.Println("|___/  \\__,_|\\___/ \\/  \\/ \\___/|_|  |_|\\_\\___|_|   ")
	fmt.Println()
	fmt.Println()
}
