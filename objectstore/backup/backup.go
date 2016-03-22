package backup

import (
	"fmt"
	"github.com/twinj/uuid"
	"io/ioutil"
	"os"
)

func SaveInsertJsons(Item []byte, namespace string, class string) {
	checkForFolder()
	fmt.Println("Saving Objects @ //JsonStack/New/POST")
	saveTempObjects(Item, 1, namespace, class)
}

func SaveUpdateJsons(Item []byte, namespace string, class string) {
	checkForFolder()
	fmt.Println("Saving Objects @ //JsonStack/New/PUT")
	saveTempObjects(Item, 2, namespace, class)
}

func SaveDeleteJsons(Item []byte, namespace string, class string) {
	checkForFolder()
	fmt.Println("Saving Objects @ //JsonStack/New/DELETE")
	saveTempObjects(Item, 3, namespace, class)
}

func checkForFolder() {
	//Create the folder to store jsons if not created
	_, errr := os.Stat("JsonStack")
	if errr != nil {
		os.Mkdir("JsonStack", 0777)
		os.Mkdir("JsonStack/New", 0777)
		os.Mkdir("JsonStack/New/POST", 0777)
		os.Mkdir("JsonStack/New/PUT", 0777)
		os.Mkdir("JsonStack/New/DELETE", 0777)
		os.Mkdir("JsonStack/Old", 0777)
		os.Mkdir("JsonStack/Old/POST", 0777)
		os.Mkdir("JsonStack/Old/PUT", 0777)
		os.Mkdir("JsonStack/Old/DELETE", 0777)
	}
}

func saveTempObjects(Item []byte, operation int, namespace string, class string) {
	greetMsg := ""
	savePath := ""
	if operation == 1 {
		greetMsg = "Commencing Saving INSERT objects...."
		savePath = "JsonStack/New/POST/"
	} else if operation == 2 {
		greetMsg = "Commencing Saving UPDATE objects...."
		savePath = "JsonStack/New/PUT/"
	} else if operation == 3 {
		greetMsg = "Commencing Saving DELETE objects...."
		savePath = "JsonStack/New/DELETE/"
	}
	fmt.Print(greetMsg)
	err := ioutil.WriteFile((savePath + getFileName(namespace, class) + ".txt"), Item, 0666)
	if err != nil {
		fmt.Println(err.Error())
	}

}

func getFileName(namespace string, class string) string {
	return (namespace + "-" + class + "-" + uuid.NewV1().String())
}
