package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Content struct {
	FileName string
	Body     string
}

func Add(v interface{}, name string) (err error) {
	name = name + ".config"
	dataset, _ := json.Marshal(v)

	file, err := os.Open(name)
	if err != nil {
		fmt.Println(err, file)
	}
	if _, err := os.Stat(name); os.IsNotExist(err) {
		_, err := os.Create(name)
		if err == nil {
			fmt.Printf("%s file created ... \n", name)
		} else {
			fmt.Printf("file cannot create please check file location ")
		}
	}

	file1, err := os.OpenFile(name, os.O_WRONLY, 0600)
	if err != nil {
		// panic(err)
		fmt.Printf("Appended into file not success please check again \n")
	}
	defer file.Close()
	if _, err = file1.WriteString(string(dataset)); err != nil {
		//panic(err)
	}
	defer file1.Close()
	return err
}

func Save(name, content string) (err error) {
	name = name + ".config"
	//dataset, _ := json.Marshal(v)

	//mt.Printf("%s file created ... \n", name)
	file, err := os.Open(name)
	if err != nil {
		fmt.Println(err, file)
	}
	if _, err := os.Stat(name); os.IsNotExist(err) {
		_, err := os.Create(name)
		if err == nil {
			fmt.Printf("%s file created ... \n", name)
		} else {
			fmt.Printf("file cannot create please check file location ")
		}
	}

	file1, err := os.OpenFile(name, os.O_WRONLY, 0600)
	if err != nil {
		// panic(err)
		fmt.Printf("Appended into file not success please check again \n")
	}
	defer file.Close()
	if _, err = file1.WriteString(content); err != nil {
		//panic(err)
	}
	defer file1.Close()
	return err
}

func Get(name string) (out []byte, err error) {
	name = name + ".config"
	file, e := ioutil.ReadFile(name)
	if e != nil {

		err = e
		out = nil
		return
	}
	err = nil
	out = file
	return
}

func GetConfigs() []string {
	files1, _ := filepath.Glob("*.config")

	return files1
}
