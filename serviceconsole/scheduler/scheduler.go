package main

import (
	"duov6.com/fws"
	"duov6.com/serviceconsole/scheduler/core"
	"duov6.com/term"
)

type Scheduler struct {
}

func (s *Scheduler) Start() {
	fws.Attach("ProcessScheduler")
	downloader := core.Downloader{}
	term.Write("Starting Serviec Console Scheduler...", term.Debug)
	downloader.Start()
}

func main() {
	scheduler := Scheduler{}
	scheduler.Start()
}
