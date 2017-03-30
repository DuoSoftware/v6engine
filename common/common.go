package common

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/twinj/uuid"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
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
			return GetHash(uuid.NewV1().String())
		}
	} else {
		return GetHash(uuid.NewV1().String())
	}
}

func ErrorJson(message string) string {
	return "{\"Error\":true,\"Message\":\"" + message + "\"}"
}

func MsgJson(message string) string {
	return "{\"Error\":false,\"Message\":\"" + message + "\"}"
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

func JWTPayload(issu, securitytoken, userid, email, domain string, b []byte) map[string]interface{} {
	payload := make(map[string]interface{})
	scope := make(map[string]interface{})
	json.Unmarshal(b, &scope)
	payload["iss"] = issu
	payload["aud"] = securitytoken
	payload["sub"] = "dwauth|" + userid
	//payload["eml"] = email
	payload["iss"] = domain
	payload["scope"] = scope
	return payload
}

func Jwt(secret string, payload map[string]interface{}) string {
	header := make(map[string]string)
	header["alg"] = "HS256"
	header["typ"] = "JWT"
	b, _ := json.Marshal(header)
	b2, _ := json.Marshal(payload)
	other := base64.StdEncoding.EncodeToString(b) + "." + base64.StdEncoding.EncodeToString(b2)
	return other + "." + ComputeHmac256(other, secret)
}

func JwtUnload(key string) map[string]interface{} {
	jwt := make(map[string]interface{})
	array := strings.Split(key, ".")
	if len(array) != 3 {
		return jwt
	}
	str := array[1]
	data, _ := base64.StdEncoding.DecodeString(str)
	strJwt := string(data)
	if len(strJwt) != (strings.LastIndex(strJwt, "}") + 1) {
		strJwt += "}"
	}
	err1 := json.Unmarshal([]byte(strJwt), &jwt)
	if err1 != nil {
		fmt.Println("jwt Error decoding " + strJwt)
		fmt.Println(err1)
	}
	return jwt
}

func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func GetProcessorUsage() (value float64) {
	if runtime.GOOS == "linux" {
		value = GetCurrentCPUusage()
	} else {
		value = 0
	}
	return
}

func VerifyConfigFiles() (config map[string]interface{}) {
	fmt.Println("Reading Environmental Variables...")

	objOsUrl := os.Getenv("OBJECTSTORE_URL")
	authOsUrl := os.Getenv("AUTH_URL")
	cebOsUrl := os.Getenv("CEB_URL")
	logOsUrl := os.Getenv("LOGSTASH_URL")

	config = make(map[string]interface{})

	content, err := ioutil.ReadFile("agent.config")
	if err == nil {
		_ = json.Unmarshal(content, &config)
	}

	if config["cebUrl"] == nil {
		config["cebUrl"] = "localhost:5000"
	}

	if cebOsUrl != "" {
		config["cebUrl"] = cebOsUrl
	}
	if objOsUrl != "" {
		config["objUrl"] = objOsUrl
	}
	if authOsUrl != "" {
		config["authUrl"] = authOsUrl
	}
	if logOsUrl != "" {
		config["logstashUrl"] = logOsUrl
	}
	config["canMonitorOutput"] = true

	byteArray, _ := json.Marshal(config)
	_ = ioutil.WriteFile("agent.config", byteArray, 0666)

	fmt.Println(config)

	return
}
