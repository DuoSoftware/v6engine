package authlib

import (
	"bytes"
	"duov6.com/common"
	//"encoding/base64"
	//"duov6.com/session"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type auth0 struct {
}

func (a *auth0) RegisterToken(object map[string]string) (AuthCertificate, string) {
	var auth AuthCertificate
	url := object["url"]
	fmt.Println("URL:>", url)
	token := object["token"]
	domain := object["domain"]
	m := common.JwtUnload(token)
	if m["sub"] == "" {
		return auth, "Error Processing jwt token."
	}
	sub, _ := m["sub"].(string)
	aud, _ := m["aud"].(string)
	iat, _ := m["iat"].(string)
	exp, _ := m["exp"].(string)
	key := common.GetHash(sub + aud + iat + exp)
	h := AuthHandler{}
	s, err := h.GetSession(key, domain)
	if err == "" {
		return s, ""
	}
	var jsonStr = []byte(`{"id_token":"` + token + `"}`)
	req, err1 := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err1 := client.Do(req)
	if err1 != nil {
		//panic(err)
		fmt.Println(err)
	}
	defer resp.Body.Close()
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	o := make(map[string]interface{})
	err2 := json.Unmarshal(body, &o)
	if err2 != nil {
		fmt.Println(err1)
		return auth, "Error Decoding request !" + err1.Error()
	}
	email, _ := o["email"].(string)
	user, status := h.GetUser(email)
	if status != "" {
		user.EmailAddress, _ = o["email"].(string)
		user.UserID, _ = m["sub"].(string)
		user.Name, _ = o["nickname"].(string)
		randText := common.RandText(5)
		user.Password = randText
		user.ConfirmPassword = randText
		user.Active = true
		user, _ = h.SaveUser(user, false, "registertoken")
	}

	auth.Email = user.EmailAddress
	auth.SecurityToken = key
	auth.Domain = domain
	auth.UserID = user.UserID
	auth.Name = user.Name
	auth.Otherdata = make(map[string]string)
	auth.Otherdata["auth0"] = token

	return auth, ""

}
