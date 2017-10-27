package main

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"duov6.com/common"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"
)

func main() {
	s6()
}

func s1() {
	nStr := "dFZLVXRjeF9uOXJ0NWFmWV8yV0ZOdlU2UGxGTWdnQ2F0c1ozbDRSakt4SDBqZ2RMcTZDU2NiMFAzWkdYWWJQelh2bW1MaVdaaXpwYi1oMHF1cDVqem5Pdk9yLURodzk5MDg1ODRCU2dDODNZYWNqV05xRUszdXJ4aHlFMmpXandSbTJOOTVXR2diNW16RTVYbVpJdmt2eVhubjdYOGR2Z0ZQRjVRd0luZ0dzREc4THlIdUpXbGFEaHJfRVBMTVc0d0h2SDB6WkN1Uk1BUklKbW1xaU15M1ZENGZ0cTRuUzVzOHZKTDBwVlNya3VOb2p0b2twODRBdGtBRENEVV9CVWhyYzJzSWdmbnZaMDNrb0NRUm9abVdpSHU4NlN1SlpZa0RGc3RWVFZTUjBoaVh1ZEZsZlEyck9oUGxwT2Jta3U2OGxYdy03Vi1QN2p3clFSRmZRVlh3"
	decN, err := base64.StdEncoding.DecodeString(nStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	n := big.NewInt(0)
	n.SetBytes(decN)

	eStr := "AQAB"
	decE, err := base64.StdEncoding.DecodeString(eStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	var eBytes []byte
	if len(decE) < 8 {
		eBytes = make([]byte, 8-len(decE), 8)
		eBytes = append(eBytes, decE...)
	} else {
		eBytes = decE
	}
	eReader := bytes.NewReader(eBytes)
	var e uint64
	err = binary.Read(eReader, binary.BigEndian, &e)
	if err != nil {
		fmt.Println(err)
		return
	}
	pKey := rsa.PublicKey{N: n, E: int(e)}

	pub, err := ssh.NewPublicKey(&pKey)
	if err != nil {
		// do something
	}

	fmt.Println(pKey)

	pubBytes := pub.Marshal()

	fmt.Println(string(pubBytes))

	pubkey, _ := x509.MarshalPKIXPublicKey(pKey)
	ioutil.WriteFile("public.key", pubkey, 0777)

	fmt.Println("-----BEGIN PUBLIC KEY-----\n" + (base64.StdEncoding.EncodeToString(pubBytes)) + "\n-----END PUBLIC KEY-----\n")
}

func s2() {
	// decode the base64 bytes for n
	nb, err := base64.RawURLEncoding.DecodeString("dFZLVXRjeF9uOXJ0NWFmWV8yV0ZOdlU2UGxGTWdnQ2F0c1ozbDRSakt4SDBqZ2RMcTZDU2NiMFAzWkdYWWJQelh2bW1MaVdaaXpwYi1oMHF1cDVqem5Pdk9yLURodzk5MDg1ODRCU2dDODNZYWNqV05xRUszdXJ4aHlFMmpXandSbTJOOTVXR2diNW16RTVYbVpJdmt2eVhubjdYOGR2Z0ZQRjVRd0luZ0dzREc4THlIdUpXbGFEaHJfRVBMTVc0d0h2SDB6WkN1Uk1BUklKbW1xaU15M1ZENGZ0cTRuUzVzOHZKTDBwVlNya3VOb2p0b2twODRBdGtBRENEVV9CVWhyYzJzSWdmbnZaMDNrb0NRUm9abVdpSHU4NlN1SlpZa0RGc3RWVFZTUjBoaVh1ZEZsZlEyck9oUGxwT2Jta3U2OGxYdy03Vi1QN2p3clFSRmZRVlh3")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(nb))

	e := 65537
	// The default exponent is usually 65537, so just compare the
	// base64 for [1,0,1] or [0,1,0,1]
	// if e != "AQAB" && e != "AAEAAQ" {
	// 	// still need to decode the big-endian int
	// 	log.Fatal("need to deocde e:", e)
	// }

	pk := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nb),
		E: e,
	}

	der, err := x509.MarshalPKIXPublicKey(pk)
	if err != nil {
		log.Fatal(err)
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	}

	var out bytes.Buffer
	pem.Encode(&out, block)
	fmt.Println(out.String())
}

func s3() {
	nStr := "dFZLVXRjeF9uOXJ0NWFmWV8yV0ZOdlU2UGxGTWdnQ2F0c1ozbDRSakt4SDBqZ2RMcTZDU2NiMFAzWkdYWWJQelh2bW1MaVdaaXpwYi1oMHF1cDVqem5Pdk9yLURodzk5MDg1ODRCU2dDODNZYWNqV05xRUszdXJ4aHlFMmpXandSbTJOOTVXR2diNW16RTVYbVpJdmt2eVhubjdYOGR2Z0ZQRjVRd0luZ0dzREc4THlIdUpXbGFEaHJfRVBMTVc0d0h2SDB6WkN1Uk1BUklKbW1xaU15M1ZENGZ0cTRuUzVzOHZKTDBwVlNya3VOb2p0b2twODRBdGtBRENEVV9CVWhyYzJzSWdmbnZaMDNrb0NRUm9abVdpSHU4NlN1SlpZa0RGc3RWVFZTUjBoaVh1ZEZsZlEyck9oUGxwT2Jta3U2OGxYdy03Vi1QN2p3clFSRmZRVlh3"
	decN, err := base64.StdEncoding.DecodeString(nStr)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(string(decN))
	}

	n := big.NewInt(0)
	n.SetBytes(decN)

	eStr := "AQAB"
	decE, err := base64.StdEncoding.DecodeString(eStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	var eBytes []byte
	if len(decE) < 8 {
		eBytes = make([]byte, 8-len(decE), 8)
		eBytes = append(eBytes, decE...)
	} else {
		eBytes = decE
	}
	eReader := bytes.NewReader(eBytes)
	var e uint64
	err = binary.Read(eReader, binary.BigEndian, &e)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("---------------")
	fmt.Println(e)
	fmt.Println("---------------")

	//pKey := rsa.PublicKey{N: n, E: int(e)}

	pk := &rsa.PublicKey{
		N: n,
		E: int(e),
	}

	der, err := x509.MarshalPKIXPublicKey(pk)
	if err != nil {
		fmt.Println("ding : " + err.Error())
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	}

	var out bytes.Buffer
	pem.Encode(&out, block)
	fmt.Println(out.String())
}

func s4() {

	auth_token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ilg1ZVhrNHh5b2pORnVtMWtsMll0djhkbE5QNC1jNTdkTzZRR1RWQndhTmsifQ.eyJleHAiOjE0OTg0NTg2NzQsIm5iZiI6MTQ5ODQ1NTA3NCwidmVyIjoiMS4wIiwiaXNzIjoiaHR0cHM6Ly9sb2dpbi5taWNyb3NvZnRvbmxpbmUuY29tLzUwYjI1MjQ2LTVjOTMtNDM2MC1hMjJiLTdmYzk0YTQ1MGE2Yi92Mi4wLyIsInN1YiI6ImM5ODlhMjJjLWU5MDctNGUxMi1hMjk0LTMyOTdjOTUzOTFjMSIsImF1ZCI6ImQ2Y2JiOTJhLWU1MGYtNGZjZS05NjJiLTRlNWI5YWE3MGFjYiIsIm5vbmNlIjoiZGVmYXVsdE5vbmNlIiwiaWF0IjoxNDk4NDU1MDc0LCJhdXRoX3RpbWUiOjE0OTg0NTUwNzQsImdpdmVuX25hbWUiOiJTbW9vdGhGbG93IiwiZmFtaWx5X25hbWUiOiJEZXYiLCJuYW1lIjoiU21vb3RoRmxvdyBEZXYiLCJpZHAiOiJmYWNlYm9vay5jb20iLCJvaWQiOiJjOTg5YTIyYy1lOTA3LTRlMTItYTI5NC0zMjk3Yzk1MzkxYzEiLCJjb3VudHJ5IjoiQXJtZW5pYSIsImV4dGVuc2lvbl9UZW5hbnQiOiJzZmRldiIsImVtYWlscyI6WyJzbW9vdGhmbG93ZGV2QGdtYWlsLmNvbSJdLCJ0ZnAiOiJCMkNfMV9TRi1TaWduSW5VcC1Qb2xpY3ktTGl2ZSJ9.oiOoSEp5NVrftfpn427AJrh-uEFJmEZuh_Mbi-oELAlhNWK6r53wWLtCkqg9nyI0gaADLAdqxiQEnq1s8_UbGveGfFK7JN3jyW9SeuI-rnFwmnn2t1KrbJlxrRUUqLZFGzjmKQkR1ZO3UMb7lpkiedrcZhxSssS_VCcgeCzi0NWeQMjmChnmuAL1OtVF48oiEUGQ09gsIbvCGBluxZv7bjIUQVVrIFQS9Z1AwVIgzALzu9f4S8TMZ7-1XcJfGeQ4uLBcTemO31021pv8obqMG8jZ0eJ4BeDGJVFTwmjiDBAb1N3zAjv-TXU-MyDLkVVo8LCGftfXcJRb7KYsoO49Cw"
	w := strings.Split(auth_token, ".")
	h_, s_ := w[0], w[2]

	if m := len(h_) % 4; m != 0 {
		h_ += strings.Repeat("=", 4-m)
	}
	if m := len(s_) % 4; m != 0 {
		s_ += strings.Repeat("=", 4-m)
	}

	nStr := common.EncodeToBase64("tVKUtcx_n9rt5afY_2WFNvU6PlFMggCatsZ3l4RjKxH0jgdLq6CScb0P3ZGXYbPzXvmmLiWZizpb-h0qup5jznOvOr-Dhw9908584BSgC83YacjWNqEK3urxhyE2jWjwRm2N95WGgb5mzE5XmZIvkvyXnn7X8dvgFPF5QwIngGsDG8LyHuJWlaDhr_EPLMW4wHvH0zZCuRMARIJmmqiMy3VD4ftq4nS5s8vJL0pVSrkuNojtokp84AtkADCDU_BUhrc2sIgfnvZ03koCQRoZmWiHu86SuJZYkDFstVTVSR0hiXudFlfQ2rOhPlpObmku68lXw-7V-P7jwrQRFfQVXw")
	decN, err := base64.URLEncoding.DecodeString(nStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	n := big.NewInt(0)
	n.SetBytes(decN)

	pKey := rsa.PublicKey{N: n, E: 65537}
	//inblockOauth := base64.URLEncoding.DecodeString(w[1])
	toHash := w[0] + "." + w[1]
	digestOauth, err := base64.URLEncoding.DecodeString(s_)

	hasherOauth := sha256.New()

	hasherOauth.Write([]byte(toHash))

	// verification of the signature
	err = rsa.VerifyPKCS1v15(&pKey, crypto.SHA256, hasherOauth.Sum(nil), digestOauth)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("verified")
	}
}

func s5() {

	auth_token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ilg1ZVhrNHh5b2pORnVtMWtsMll0djhkbE5QNC1jNTdkTzZRR1RWQndhTmsifQ.eyJleHAiOjE0OTg0NTg2NzQsIm5iZiI6MTQ5ODQ1NTA3NCwidmVyIjoiMS4wIiwiaXNzIjoiaHR0cHM6Ly9sb2dpbi5taWNyb3NvZnRvbmxpbmUuY29tLzUwYjI1MjQ2LTVjOTMtNDM2MC1hMjJiLTdmYzk0YTQ1MGE2Yi92Mi4wLyIsInN1YiI6ImM5ODlhMjJjLWU5MDctNGUxMi1hMjk0LTMyOTdjOTUzOTFjMSIsImF1ZCI6ImQ2Y2JiOTJhLWU1MGYtNGZjZS05NjJiLTRlNWI5YWE3MGFjYiIsIm5vbmNlIjoiZGVmYXVsdE5vbmNlIiwiaWF0IjoxNDk4NDU1MDc0LCJhdXRoX3RpbWUiOjE0OTg0NTUwNzQsImdpdmVuX25hbWUiOiJTbW9vdGhGbG93IiwiZmFtaWx5X25hbWUiOiJEZXYiLCJuYW1lIjoiU21vb3RoRmxvdyBEZXYiLCJpZHAiOiJmYWNlYm9vay5jb20iLCJvaWQiOiJjOTg5YTIyYy1lOTA3LTRlMTItYTI5NC0zMjk3Yzk1MzkxYzEiLCJjb3VudHJ5IjoiQXJtZW5pYSIsImV4dGVuc2lvbl9UZW5hbnQiOiJzZmRldiIsImVtYWlscyI6WyJzbW9vdGhmbG93ZGV2QGdtYWlsLmNvbSJdLCJ0ZnAiOiJCMkNfMV9TRi1TaWduSW5VcC1Qb2xpY3ktTGl2ZSJ9.oiOoSEp5NVrftfpn427AJrh-uEFJmEZuh_Mbi-oELAlhNWK6r53wWLtCkqg9nyI0gaADLAdqxiQEnq1s8_UbGveGfFK7JN3jyW9SeuI-rnFwmnn2t1KrbJlxrRUUqLZFGzjmKQkR1ZO3UMb7lpkiedrcZhxSssS_VCcgeCzi0NWeQMjmChnmuAL1OtVF48oiEUGQ09gsIbvCGBluxZv7bjIUQVVrIFQS9Z1AwVIgzALzu9f4S8TMZ7-1XcJfGeQ4uLBcTemO31021pv8obqMG8jZ0eJ4BeDGJVFTwmjiDBAb1N3zAjv-TXU-MyDLkVVo8LCGftfXcJRb7KYsoO49Cw"
	w := strings.Split(auth_token, ".")
	h_, s_ := w[0], w[2]

	if m := len(h_) % 4; m != 0 {
		h_ += strings.Repeat("=", 4-m)
	}
	if m := len(s_) % 4; m != 0 {
		s_ += strings.Repeat("=", 4-m)
	}

	var err error

	decN := []byte("tVKUtcx_n9rt5afY_2WFNvU6PlFMggCatsZ3l4RjKxH0jgdLq6CScb0P3ZGXYbPzXvmmLiWZizpb-h0qup5jznOvOr-Dhw9908584BSgC83YacjWNqEK3urxhyE2jWjwRm2N95WGgb5mzE5XmZIvkvyXnn7X8dvgFPF5QwIngGsDG8LyHuJWlaDhr_EPLMW4wHvH0zZCuRMARIJmmqiMy3VD4ftq4nS5s8vJL0pVSrkuNojtokp84AtkADCDU_BUhrc2sIgfnvZ03koCQRoZmWiHu86SuJZYkDFstVTVSR0hiXudFlfQ2rOhPlpObmku68lXw-7V-P7jwrQRFfQVXw")

	n := big.NewInt(0)
	n.SetBytes(decN)

	pKey := &rsa.PublicKey{N: n, E: 65537}

	// verification of the signature
	err = Verify(auth_token, pKey)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("verified")
	}

}

func s6() {

	auth_token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ilg1ZVhrNHh5b2pORnVtMWtsMll0djhkbE5QNC1jNTdkTzZRR1RWQndhTmsifQ.eyJleHAiOjE0OTg0NTg2NzQsIm5iZiI6MTQ5ODQ1NTA3NCwidmVyIjoiMS4wIiwiaXNzIjoiaHR0cHM6Ly9sb2dpbi5taWNyb3NvZnRvbmxpbmUuY29tLzUwYjI1MjQ2LTVjOTMtNDM2MC1hMjJiLTdmYzk0YTQ1MGE2Yi92Mi4wLyIsInN1YiI6ImM5ODlhMjJjLWU5MDctNGUxMi1hMjk0LTMyOTdjOTUzOTFjMSIsImF1ZCI6ImQ2Y2JiOTJhLWU1MGYtNGZjZS05NjJiLTRlNWI5YWE3MGFjYiIsIm5vbmNlIjoiZGVmYXVsdE5vbmNlIiwiaWF0IjoxNDk4NDU1MDc0LCJhdXRoX3RpbWUiOjE0OTg0NTUwNzQsImdpdmVuX25hbWUiOiJTbW9vdGhGbG93IiwiZmFtaWx5X25hbWUiOiJEZXYiLCJuYW1lIjoiU21vb3RoRmxvdyBEZXYiLCJpZHAiOiJmYWNlYm9vay5jb20iLCJvaWQiOiJjOTg5YTIyYy1lOTA3LTRlMTItYTI5NC0zMjk3Yzk1MzkxYzEiLCJjb3VudHJ5IjoiQXJtZW5pYSIsImV4dGVuc2lvbl9UZW5hbnQiOiJzZmRldiIsImVtYWlscyI6WyJzbW9vdGhmbG93ZGV2QGdtYWlsLmNvbSJdLCJ0ZnAiOiJCMkNfMV9TRi1TaWduSW5VcC1Qb2xpY3ktTGl2ZSJ9.oiOoSEp5NVrftfpn427AJrh-uEFJmEZuh_Mbi-oELAlhNWK6r53wWLtCkqg9nyI0gaADLAdqxiQEnq1s8_UbGveGfFK7JN3jyW9SeuI-rnFwmnn2t1KrbJlxrRUUqLZFGzjmKQkR1ZO3UMb7lpkiedrcZhxSssS_VCcgeCzi0NWeQMjmChnmuAL1OtVF48oiEUGQ09gsIbvCGBluxZv7bjIUQVVrIFQS9Z1AwVIgzALzu9f4S8TMZ7-1XcJfGeQ4uLBcTemO31021pv8obqMG8jZ0eJ4BeDGJVFTwmjiDBAb1N3zAjv-TXU-MyDLkVVo8LCGftfXcJRb7KYsoO49Cw"
	w := strings.Split(auth_token, ".")
	h_, s_ := w[0], w[2]

	if m := len(h_) % 4; m != 0 {
		h_ += strings.Repeat("=", 4-m)
	}
	if m := len(s_) % 4; m != 0 {
		s_ += strings.Repeat("=", 4-m)
	}

	//var err error

	//	decN := []byte("tVKUtcx_n9rt5afY_2WFNvU6PlFMggCatsZ3l4RjKxH0jgdLq6CScb0P3ZGXYbPzXvmmLiWZizpb-h0qup5jznOvOr-Dhw9908584BSgC83YacjWNqEK3urxhyE2jWjwRm2N95WGgb5mzE5XmZIvkvyXnn7X8dvgFPF5QwIngGsDG8LyHuJWlaDhr_EPLMW4wHvH0zZCuRMARIJmmqiMy3VD4ftq4nS5s8vJL0pVSrkuNojtokp84AtkADCDU_BUhrc2sIgfnvZ03koCQRoZmWiHu86SuJZYkDFstVTVSR0hiXudFlfQ2rOhPlpObmku68lXw-7V-P7jwrQRFfQVXw")
	decN, _ := json.Marshal("tVKUtcx_n9rt5afY_2WFNvU6PlFMggCatsZ3l4RjKxH0jgdLq6CScb0P3ZGXYbPzXvmmLiWZizpb-h0qup5jznOvOr-Dhw9908584BSgC83YacjWNqEK3urxhyE2jWjwRm2N95WGgb5mzE5XmZIvkvyXnn7X8dvgFPF5QwIngGsDG8LyHuJWlaDhr_EPLMW4wHvH0zZCuRMARIJmmqiMy3VD4ftq4nS5s8vJL0pVSrkuNojtokp84AtkADCDU_BUhrc2sIgfnvZ03koCQRoZmWiHu86SuJZYkDFstVTVSR0hiXudFlfQ2rOhPlpObmku68lXw-7V-P7jwrQRFfQVXw")
	n := big.NewInt(0)
	n.SetBytes(decN)

	pKey := rsa.PublicKey{N: n, E: 65537}

	savePublicPEMKey("gg.pem", pKey)

}

func Verify(token string, key *rsa.PublicKey) error {

	fmt.Println(key)

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errors.New("jws: invalid token received, token must have 3 parts")
	}

	signedContent := parts[0] + "." + parts[1]
	signatureString, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return err
	}

	h := sha256.New()
	h.Write([]byte(signedContent))

	return rsa.VerifyPKCS1v15(key, crypto.SHA256, h.Sum(nil), signatureString)
}

func savePublicPEMKey(fileName string, pubkey rsa.PublicKey) {
	asn1Bytes, err := asn1.Marshal(pubkey)

	_ = asn1.TagOctetString
	//asn1Bytes, err := x509.MarshalPKIXPublicKey(&pubkey)
	checkError(err)

	var pemkey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	pemfile, err := os.Create(fileName)
	checkError(err)
	defer pemfile.Close()

	err = pem.Encode(pemfile, pemkey)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
