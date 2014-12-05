package main

import (
	"duov6.com/common"
	"duov6.com/stat"
	"duov6.com/term"
	"time"
)

func main() {

	term.GetConfig()
	term.Write(s.Format("20060102"), term.Debug)
	term.Write("Lable", term.Debug)
	stat.Start()
	go ErrorMethods()
	go Informaton()
	term.StartCommandLine()
}

func ErrorMethods() {
	count := 10
	for i := 0; i < count; i++ {
		stats := stat.States{}
		stats.MethodName = common.RandText(10)
		stats.NameSpace = "Lasitha"
		stats.ObjectSize = 10
		stats.ElapsedTime = 1
		stats.Status = stat.Error
		stat.Add(stats)
	}
}

func Informaton() {
	count := 10
	for i := 0; i < count; i++ {
		stats := stat.States{}
		stats.MethodName = common.RandText(10)
		stats.NameSpace = "Lasitha"
		stats.ObjectSize = 10
		stats.ElapsedTime = 1
		stats.Status = stat.Sucess
		stat.Add(stats)
	}
}
