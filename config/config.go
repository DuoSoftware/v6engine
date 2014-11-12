package config

import (
	"bufio"
	"fmt"
	"os"
	//"strconv"
	"strings"
)

type File struct {
	m        map[string]string
	Filename string
}

//Config.txt file load for modifications
func (c *File) loadfile() {
	configFile := c.Filename
	if c.m == nil {
		c.m = make(map[string]string)
	}
	//file open
	file, err := os.Open(configFile)
	if file != nil {
		fmt.Println(file)
	} else {
		fmt.Println("File cannot Open", file, err)
	}
	//read file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		//split key and value
		if scanner.Text() != "" {
			stringSlice := strings.Split(scanner.Text(), "---->")
			c.m[stringSlice[0]] = stringSlice[1]
		}
	}
}

//add new key and value if key exist value will update
func (c *File) Add(Key, Value string) {
	if c.m == nil {
		c.m = make(map[string]string)
	}
	c.m[Key] = Value
	fmt.Println(Key, Value)
}

//return value to given key
func (c *File) Get(Key string) string {
	if c.m == nil {
		c.m = make(map[string]string)
	}
	result := c.m[Key]
	if result == "" {
		fmt.Println("Key Not Found ")
	} else {
		fmt.Println(c.m[Key], "is the value of .", Key)
	}
	return c.m[Key]
}

//delete data relevent to given key from config.txt
func (c *File) Delete(Key string) {
	if c.m == nil {
		c.m = make(map[string]string)
	}
	delete(c.m, Key)

}

//write modified data to config.txt
func (c *File) writetoFile() {
	configFile := c.Filename

	if c.m == nil {
		c.m = make(map[string]string)
	}

	file, err := os.Open(configFile)

	if err != nil {
		fmt.Println(err, file)
	}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		_, err := os.Create(configFile)
		if err == nil {
			fmt.Printf("%s file  created ... \n", configFile)
		} else {
			fmt.Printf("file cannot create please check file location ")
		}
	}

	file1, err := os.OpenFile(configFile, os.O_WRONLY, 0600)
	if err != nil {
		// panic(err)
		fmt.Printf("Appended into file not success please check again \n")
	}
	defer file.Close()
	//fmt.Println(c.m,"before write to file 23")
	for k, v := range c.m {
		//fmt.Println(k,v,"\n")
		if _, err = file1.WriteString(k + "---->" + "<" + v + ">\n"); err != nil {
			panic(err)
		}
	}
}

func func_name() {

}
