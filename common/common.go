package common

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/twinj/uuid"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandText(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GetGUID() string {
	if runtime.GOOS == "linux" {
		//h.Write()
		out, err := exec.Command("uuidgen").Output()
		h := md5.New()
		h.Write(out)
		if err == nil {
			return hex.EncodeToString(h.Sum(nil))
		} else {
			return uuid.NewV1().String()
		}
	} else {
		return uuid.NewV1().String()
	}
}

func GetHash(input string) string {
	h := md5.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

func RandomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func SaveFile(fileName, Text string) (err error) {

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		_, err := os.Create(fileName)
		if err == nil {
			fmt.Printf("%s file created ... \n", fileName)
		} else {
			fmt.Printf("file cannot create please check file location ")
		}
	}
	//os.OP
	file1, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		// panic(err)
		fmt.Printf("Appended into file not success please check again \n")
	}
	//defer file.Close()
	if _, err = file1.WriteString(string(Text)); err != nil {
		fmt.Printf("Failed to write log \n" + err.Error())
		//panic(err)
	}
	defer file1.Close()
	return err

}

func EncodeToBase64(message string) (retour string) {

	base64Byte := make([]byte, base64.StdEncoding.EncodedLen(len(message)))

	base64.StdEncoding.Encode(base64Byte, []byte(message))

	return string(base64Byte)

}

func DecodeFromBase64(message string) (retour string) {

	base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(message)))

	base64.StdEncoding.Decode(base64Text, []byte(message))

	return string(base64Text)

}

func PublishLog(fileName string, Body string) {

	if runtime.GOOS == "linux" {
		date := string(time.Now().Local().Format("2006-01-02 @ 15:04:05"))
		_, _ = exec.Command("sh", "-c", "echo "+date+" >> "+fileName).Output()
		_, _ = exec.Command("sh", "-c", "echo "+Body+" >> "+fileName).Output()
	} else {
		ff, err := os.OpenFile(fileName, os.O_APPEND, 0666)
		if err != nil {
			ff, err = os.Create(fileName)
			ff, err = os.OpenFile(fileName, os.O_APPEND, 0666)
		}
		_, err = ff.Write([]byte(string(time.Now().Local().Format("2006-01-02 @ 15:04:05")) + "  "))
		_, err = ff.Write([]byte(Body))
		_, err = ff.Write([]byte("\r\n"))
		if err != nil {
			fmt.Println(err.Error())
		}

		ff.Close()
	}

}
