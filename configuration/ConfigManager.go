package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Node struct {
	Id       string  `json:"-"`
	ParentId string  `json:"-"`
	Name     string  `json:"name"`
	Value    string  `json:"Value,omitempty"`
	Children []*Node `json:"children,omitempty"`
}

func (this *Node) Size() int {
	var size int = len(this.Children)
	for _, c := range this.Children {
		size += c.Size()
	}
	return size
}

func (this *Node) Add(nodes ...*Node) bool {
	var size = this.Size()
	for _, n := range nodes {
		if n.ParentId == this.Id {
			this.Children = append(this.Children, n)
		} else {
			for _, c := range this.Children {
				if c.Add(n) {
					break
				}
			}
		}
	}
	return this.Size() == size+len(nodes)
}

func (this *Node) Get(name string) (node *Node) {
	for index, element := range this.Children {
		if element.Name == name {
			index = index
			node = element
			return
		}
	}
	return
}

func (this *Node) Find(key string, value string) (node *Node) {

	for index, element := range this.Children {
		if element.Name == key && element.Value == value {
			index = index
			node = element
			break
		}
	}

	return
}

func SaveConfigTree(obj interface{}) {
	dataset, _ := json.Marshal(obj)
	configFile := "Config.json"
	file, err := os.Open(configFile)
	if err != nil {
		fmt.Println(err, file)
	}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		_, err := os.Create(configFile)
		if err == nil {
			fmt.Printf("%s file created ... \n", configFile)
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
	if _, err = file1.WriteString(string(dataset)); err != nil {
		panic(err)
	}
}

func LoadConfigTree(jsonfile string) (outObj Node) {
	file, e := ioutil.ReadFile(jsonfile)
	if e != nil {
		//fmt.Printf("File error: %v\n", e)
		//os.Exit(1)
		if _, err := os.Stat(jsonfile); os.IsNotExist(err) {
			_, err := os.Create(jsonfile)
			if err == nil {
				fmt.Printf("%s file created ... \n", jsonfile)
			} else {
				fmt.Printf("file cannot create please check file location ")
			}
		}
	}

	json.Unmarshal(file, &outObj)
	return
}

func Save(jsonfile string, config map[string]interface{}) {
	//fmt.Println(config,jsonfile)
	
	dataset, _ := json.Marshal(config)
	file, err := os.Open(jsonfile)
	if err != nil {
		fmt.Println(err, file)
	}
	if _, err := os.Stat(jsonfile); os.IsNotExist(err) {
		_, err := os.Create(jsonfile)
		if err == nil {
			fmt.Printf("%s file created ... \n", jsonfile)
		} else {
			fmt.Printf("file cannot create please check file location ")
		}
	}
	file1, err := os.OpenFile(jsonfile, os.O_WRONLY, 0600)
	if err != nil {
		// panic(err)
		fmt.Printf("Appended into file not success please check again \n")
	}
	defer file.Close()

	if _, err = file1.WriteString(string(dataset)); err != nil {
		panic(err)
	}
	fmt.Println(string(dataset))


}

func Load(jsonfile string) (map[string]interface{}) {

	file, e := ioutil.ReadFile(jsonfile)
	if e != nil {
		if _, err := os.Stat(jsonfile); os.IsNotExist(err) {
			_, err := os.Create(jsonfile)
			if err == nil {
				fmt.Printf("%s file created ... \n", jsonfile)
			} else {
				fmt.Printf("file cannot create please check file location ")
			}
		}
	}
	var data map[string]interface{}
	err := json.Unmarshal(file, &data)
	if err!=nil{
		panic(err)
	}

	//fmt.Println(data)

	//var jsontype Node
    //json.Unmarshal(file, &jsontype)
	return data
}

func main() {
	/*var root *Node = &Node{"001", "", "DbName", "MySql", nil}
	data := []*Node{
		&Node{"002", "001", "Db", "DuoV6", nil},
		&Node{"003", "002", "Username", "Duov6", nil},
		&Node{"004", "002", "Password", "123", nil},
		&Node{"005", "004", "tables", "Auth", nil},
		&Node{"006", "004", "tables", "Config", nil},
		&Node{"007", "004", "tables", "derective", nil},
		&Node{"008", "004", "tables", "Users", nil},
		&Node{"009", "004", "tables", "Canves", nil},
		&Node{"010", "004", "tables", "RabbitMQ", nil},
		&Node{"011", "004", "tables", "test table 2 ", nil},
		&Node{"012", "004", "tables", "test table 3", nil},
	}*/

	//fmt.Println(root.Add(data...), root.Size())
	//bytes, _ := json.MarshalIndent(root, "", "\t") //formated output
	//fmt.Println(string(bytes))

	//Save(bytes)

	node := LoadConfigTree("Config.json")
	value := node.Get("Db")
	fmt.Println(value)
////////////////////////////////////////////////////////////
	file, e := ioutil.ReadFile("Config.json")
	if e != nil {
		//fmt.Printf("File error: %v\n", e)
		//os.Exit(1)
		if _, err := os.Stat("Config.json"); os.IsNotExist(err) {
			_, err := os.Create("Config.json")
			if err == nil {
				fmt.Printf("%s file created ... \n", "Config.json")
			} else {
				fmt.Printf("file cannot create please check file location ")
			}
		}
	}

	if file !=nil{

	}

	/*var data interface{}
    json.Unmarshal(file, &data)

   msg := data.(map[string]interface{})
Save("Config.json",msg)*/


	//node1:=Node{}
	//Save("sdfghjk.json",root)
	/*fmt.Printf("Results:%v\n", value)

	var testMap map[string]interface{}
	testMap = make(map[string]interface{})
	testMap["fsdfsdf"] = 123

	Save("testconfig.json", testMap)
*/

fmt.Println(Load("testconfig.json"))

}
