package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

func checkError(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

func main() {
	cmd := exec.Command("ls", "-R")

	stdout, err := cmd.StdoutPipe()
	checkError(err)
	stderr, err := cmd.StderrPipe()
	checkError(err)

	err = cmd.Start()
	checkError(err)

	defer cmd.Wait() // Doesn't block

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	fmt.Printf("Do other stuff here! No need to wait.\n\n")

}
