package api

import (
	"duov6.com/common"
	"duov6.com/objectstore/client"
	"duov6.com/term"
	"encoding/json"
	"time"
)

type TokenManager struct {
}

func (r *TokenManager) Generate(o map[string]interface{}) string {
	o["id"] = common.RandText(10)
	nowTime := time.Now().UTC()
	o["iat"] = nowTime.Format("2006-01-02 15:04:05")
	client.Go("ignore", "com.duosoftware.auth", "tokens").StoreObject().WithKeyField("id").AndStoreOne(o).Ok()
	return o["id"].(string)
}

func (r *TokenManager) Get(requestCode string) map[string]interface{} {
	o := make(map[string]interface{})
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "tokens").GetOne().ByUniqueKey(requestCode).Ok() // fetech user autherized
	term.Write("Get Request Code : "+requestCode, term.Debug)
	if err == "" {
		if bytes != nil {
			_ = json.Unmarshal(bytes, &o)
		}
	}
	return o
}

func (r *TokenManager) Delete(id string) {
	o := make(map[string]interface{})
	o["id"] = id
	client.Go("ignore", "com.duosoftware.auth", "tmprequestcodes").DeleteObject().WithKeyField("id").AndDeleteObject(o).Ok()
}
