package main

import (
	"duov6.com/objectstore/endpoints"
	"duov6.com/objectstore/unittesting"
	"duov6.com/term"
)

func main() {
	var isUnitTestMode bool = false

	if isUnitTestMode {
		unittesting.Start()
	} else {
		splash()
		initialize()
	}
}

func initialize() {
	httpServer := endpoints.HTTPService{}
	httpServer.Start()
}

func splash() {

	term.SplashScreen("splash.art")

}
