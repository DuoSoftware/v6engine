package main

import (
	"fmt"
	"sync"
	"time"
)

var m = map[string]int{"a": 1}
var lock = sync.RWMutex{}

func main() {
	Read()
	time.Sleep(1 * time.Second)
	Write()
	fmt.Println(m)
}

func Read() {

	lock.RLock()
	defer lock.RUnlock()
	_ = m["a"]

}

func Write() {

	lock.Lock()
	defer lock.Unlock()
	m["b"] = 2

}
