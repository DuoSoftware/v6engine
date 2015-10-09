package documentcache

import (
	"fmt"
)

func Store(key string, ttl int, data interface{}) (status bool) {
	fmt.Println("Executing Store-Cache!")
	verifyDirectory()
	status = writeToFile(key, ttl, data)
	return
}

func Fetch(key string) (data interface{}) {
	fmt.Println("Executing Fetch-Cache!")
	verifyDirectory()
	body, status := readFromFile(key)
	if status {
		data = body
	} else {
		data = nil
	}
	return
}
