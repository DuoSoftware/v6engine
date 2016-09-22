package keygenerator

import (
	"duov6.com/objectstore/configuration"
	"duov6.com/objectstore/messaging"
	"fmt"
	"strconv"
	"strings"
)

func UpdateKeysInDB() {
	//Get All keys

	keys := GetAllKeyGenKeys()

	if len(keys) == 0 {
		return
	}

	for _, key := range keys {
		repos := GetRepositories(key)
		namespace, class := GetDomainClassFromKey(key)

		request := messaging.ObjectRequest{}
		request.Controls.Namespace = namespace
		request.Controls.Class = class
		request.Configuration = configuration.ConfigurationManager{}.Get("ignore", namespace, class)

		host := request.Configuration.ServerConfiguration["REDIS"]["Host"]
		port := request.Configuration.ServerConfiguration["REDIS"]["Port"]

		client, err := GetConnectionTCP(host, port)

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		for _, repo := range repos {
			count, err := strconv.Atoi(ReadKeyGenKey(&request, client))
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			_ = VerifyMaxFromDB(&request, repo, count)
		}
	}

}

func GetAllKeyGenKeys() (keys []string) {
	pattern := "KeyGenKey*"

	host := configuration.ConfigurationManager{}.Get("ignore", "ignore", "ignore").ServerConfiguration["REDIS"]["Host"]
	port := configuration.ConfigurationManager{}.Get("ignore", "ignore", "ignore").ServerConfiguration["REDIS"]["Port"]

	client, _ := GetConnectionTCP(host, port)
	keys, _ = client.Keys(pattern)
	return
}

func GetRepositories(key string) (repos []string) {
	namespace, class := GetDomainClassFromKey(key)
	configuration := configuration.ConfigurationManager{}.Get("ignore", namespace, class).StoreConfiguration["INSERT-SINGLE"]

	for _, repo := range configuration {
		repos = append(repos, repo)
	}
	return
}

func GetDomainClassFromKey(key string) (namespace string, class string) {
	value := key
	value = strings.TrimPrefix(value, "KeyGenKey.")
	tokens := strings.Split(value, ".")
	class = tokens[len(tokens)-1]
	for x := 0; x < (len(tokens) - 1); x++ {
		namespace += tokens[x] + "."
	}
	namespace = strings.TrimSuffix(namespace, ".")
	return
}
