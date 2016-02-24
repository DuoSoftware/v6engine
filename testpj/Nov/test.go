package main

import (
	//"duov6.com/queryparser"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

func main() {
	//fmt.Println(queryparser.GetDataStoreQuery("SELECT name, Id, age from Student s2 , Game g2 where age >= 10 and course = 'SLIIT' order by Id ASC;"))
	//fmt.Println(queryparser.GetElasticQuery("SELECT tttt, Id, age from Student s1, game g2 where ddf = 5@00 AND kappa not between 50 AND 100 AND kat not in ('asdf', 'wert') OR Country LIKE '%land%' order by ppds, ggg desc;", "com.duoworld.com", "test"))
	//fmt.Println(queryparser.GetElasticQuery("SELECT tttt, Id, age from Student s1, game g2;", "com.duoworld.com", "test"))
	//fmt.Println(queryparser.GetDataStoreQuery("SELECT tttt, Id, age from Student s1, game g2 order by age desc, name ;", "com.duoworld.com", "test"))
	//fmt.Println(queryparser.GetElasticQuery("select * from chatmessages where (too='usr1' AND frm='usr2') OR (too= 'usr2' AND frm='usr1');", "com.duoworld.com", "test"))

	//fmt.Println(queryparser.GetDataStoreQuery("SELECT tttt, Id, age from Student s1, game g2 where value = '50 0' AND kappa not between 50 AND 100 AND kat not in ('asdf', 'wert') order by ppds, ggg desc;", "com.duoworld.com", "test"))
	//fmt.Println(queryparser.GetElasticQuery("select * from product12thdoor where ProductCode like '%MAG%';", "com.duoworld.com", "test"))

	//fmt.Println(queryparser.GetElasticQuery("SELECT tttt, Id, age from Student s1, game g2 where Country LIKE '%land%' ;", "com.duoworld.com", "test"))
	fmt.Println(jwt())
}

func jwt() string {
	secret := "123"

	payload := make(map[string]interface{})
	scope := make(map[string]interface{})
	scope["test"] = ""

	payload["sub"] = ""
	payload["scpe"] = scope
	payload["sssss"] = "yoman"
	payload["sssss1"] = "123"
	//hashkey := other + "." + secret
	fmt.Println(payload["ssssaaas"])
	if payload["sssss1"] == "123" {
		fmt.Println(payload["sssss"])
	}
	return jwt1(secret, payload)

}

func jwt1(secret string, payload map[string]interface{}) string {
	header := make(map[string]string)
	header["alg"] = "HS256"
	header["typ"] = "JWT"
	b, _ := json.Marshal(header)
	b2, _ := json.Marshal(payload)
	other := base64.StdEncoding.EncodeToString(b) + "." + base64.StdEncoding.EncodeToString(b2)
	return other + "." + ComputeHmac256(other, secret)
}

func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
