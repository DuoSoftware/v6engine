package authlib

import (
	"duov6.com/common"
	//email "duov6.com/duonotifier/client"
	"duov6.com/objectstore/client"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"time"
)

type tempRequestGenerator struct {
}

func (r *tempRequestGenerator) GenerateRequestCode(o map[string]string) string {
	o["id"] = common.RandText(15)
	nowTime := time.Now().UTC()
	nowTime = nowTime.Add(5 * time.Minute)
	o["iat"] = nowTime.Format("2006-01-02 15:04:05")
	term.Write(o, term.Debug)
	data := make(map[string]interface{})
	for key, value := range o {
		data[key] = value
	}
	client.Go("ignore", "com.duosoftware.auth", "tmprequestcodes").StoreObject().WithKeyField("id").AndStoreOne(data).Ok()
	return o["id"]
}

func (r *tempRequestGenerator) GetRequestCode(requestCode string) (map[string]string, string) {
	o := make(map[string]string)
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "tmprequestcodes").GetOne().ByUniqueKey(requestCode).Ok() // fetech user autherized
	term.Write("GetRequestCode "+requestCode+"  ", term.Debug)
	if err == "" {
		if bytes != nil {
			//var uList LoginSessions
			data := make(map[string]interface{})
			err := json.Unmarshal(bytes, &data)
			if err == nil {
				for key, value := range data {
					term.Write("GetRequestCode "+requestCode+"  "+key, term.Debug)
					if str, ok := value.(string); ok {
						/* act on str */
						o[key] = str
					}

				}
				//Ttime1, _ := time.Parse("2006-01-02 15:04:05", o["expairyTime"])
				//Ttime2 := time.Now().UTC()
				return o, ""
			} else {
				term.Write("GetRequestCode err "+err.Error(), term.Error)
			}
		}
	} else {
		term.Write("GetRequestCode err "+err, term.Error)
	}

	return o, "Incorrect Request Code."
}

func (r *tempRequestGenerator) Remove(o map[string]interface{}) {
	term.Write(o, term.Debug)
	client.Go("ignore", "com.duosoftware.auth", "tmprequestcodes").DeleteObject().WithKeyField("id").AndDeleteOne(o).Ok()
}

func (r *tempRequestGenerator) RemoveByEmail(email, tid string) {
	fmt.Println("Removing all request tokens for email : " + email)
	for _, value := range r.GetTmpCodesByEmail(email, tid) {
		fmt.Println(value["id"])
		client.Go("ignore", "com.duosoftware.auth", "tmprequestcodes").DeleteObject().WithKeyField("id").AndDeleteOne(value).Ok()
	}
}

func (r *tempRequestGenerator) GetTmpCodesByEmail(Email, tid string) (codes []map[string]string) {
	codes = make([]map[string]string, 0)
	//("Select * From usersubscriptionreq321 where Email ='" + u.Email + "'")
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "tmprequestcodes").GetMany().ByQuerying("Select * From tmprequestcodes where Email ='" + Email + "' AND TenantID='" + tid + "';").Ok()
	if err == "" {
		if bytes != nil {
			_ = json.Unmarshal(bytes, &codes)
		}
	}
	return
}
