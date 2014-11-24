package common

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"os/exec"
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

	//h.Write()
	out, err := exec.Command("uuidgen").Output()
	h := md5.New()
	h.Write(out)
	if err == nil {
		return hex.EncodeToString(h.Sum(nil))
	} else {
		return ""
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
