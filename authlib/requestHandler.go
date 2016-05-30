package authlib

import (
	"duov6.com/common"
	//email "duov6.com/duonotifier/client"
	"duov6.com/objectstore/client"
	"duov6.com/term"
	"encoding/json"
	"time"
)

type requestHandler struct {
}

func (r *requestHandler) GenerateRequestCode(o map[string]string) string {
	o["id"] = common.RandText(5)
	nowTime := time.Now().UTC()
	nowTime = nowTime.Add(5 * time.Minute)
	o["expairyTime"] = nowTime.Format("2006-01-02 15:04:05")
	term.Write(o, term.Debug)
	client.Go("ignore", "com.duosoftware.auth", "requestcodes").StoreObject().WithKeyField("id").AndStoreOne(o).Ok()
	return o["id"]
}

func (r *requestHandler) GetRequestCode(requestCode string) (map[string]string, string) {
	o := make(map[string]string)
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "requestcodes").GetOne().ByUniqueKey(requestCode).Ok() // fetech user autherized
	term.Write("GetRequestCode "+requestCode+"  ", term.Debug)
	if err == "" {
		if bytes != nil {
			//var uList LoginSessions
			err := json.Unmarshal(bytes, &o)
			if err == nil {
				Ttime1, _ := time.Parse("2006-01-02 15:04:05", o["expairyTime"])
				Ttime2 := time.Now().UTC()
				difference := Ttime1.Sub(Ttime2)
				minutesTime := difference.Minutes()
				if minutesTime <= 0 {
					r.Remove(o)
					return o, "Expired."
				} else {
					return o, ""
				}
			} else {
				term.Write("GetRequestCode err "+err.Error(), term.Error)
			}
		}
	} else {
		term.Write("GetRequestCode err "+err, term.Error)
	}

	return o, "Error Finding And processing."
}

func (r *requestHandler) Remove(o map[string]string) {
	client.Go("ignore", "com.duosoftware.auth", "requestcodes").DeleteObject().WithKeyField("id").AndDeleteObject(o).Ok()
}
