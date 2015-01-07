package fws

import (
	"duov6.com/term"
	"fmt"
)

var logger *FWSLogger

func Switch(from string, name string, data map[string]interface{}, resources map[string]interface{}) {

	if logger == nil {
		logger = &FWSLogger{}
	}

	fmt.Println(data)

	var attrib = data["state"].(string)

	if attrib == "on" {
		fmt.Println("LOG Monitor Turned ON")
		term.AddPlugin(logger)

	} else if attrib == "off" {
		fmt.Println("LOG Monitor Turned OFF")
		term.RemovePlugin(logger)
	}
}
