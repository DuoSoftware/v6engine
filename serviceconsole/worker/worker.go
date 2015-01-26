package main

import (
	"duov6.com/serviceconsole/scheduler/core"
	"fmt"
)

type Worker struct {
}

func (w *Worker) Start() {
	fws.Attach("ProcessWorker")
	downloader := core.Downloader{}
	fmt.Println("worker start ")
	downloader.Start()
}

func main() {
	worker := Worker{}
	worker.Start()
}
