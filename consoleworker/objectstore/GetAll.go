package objectstore

import (
	"duov6.com/consoleworker/common"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func GetAll(domain string, class string) (data []map[string]interface{}, err error) {

	securityToken := "ignore"

	config := common.GetConfigurations()
	url := config["SVC_OS_URL"].(string) + domain + "/" + class + "?skip=0&take=1200000"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("securityToken", securityToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		} else {
			_ = json.Unmarshal(body, &data)
			if len(data) <= 0 {
				data = make([]map[string]interface{}, 0)
				return data, err
			} else {
				return data, nil
			}
		}
	}

	defer resp.Body.Close()

	return
}
