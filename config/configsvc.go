package config

import (
	"duov6.com/gorest"
	//"fmt"
)

type ConfigSvc struct {
	gorest.RestService
	files gorest.EndPoint `method:"GET" path:"/Config/Files" output:"[]string"`
	get   gorest.EndPoint `method:"GET" path:"/Config/Get/{filename:string}" output:"string"`
	save  gorest.EndPoint `method:"POST" path:"/Config/Save/" postdata:"Content"`
}

func (A ConfigSvc) Files() []string {
	//A.ResponseBuilder().AddHeader("Access-Control-Allow-Origin", "*")
	return GetConfigs()
}

func (A ConfigSvc) Get(filename string) (s string) {
	//A.ResponseBuilder().AddHeader("Access-Control-Allow-Origin", "*")

	b, err := Get(filename)
	if err == nil {
		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
		return
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(err.Error()))
		return
	}
}

func (A ConfigSvc) Save(c Content) {
	//A.ResponseBuilder().AddHeader("Access-Control-Allow-Origin", "*")
	//fmt.Println(A.Context.Request().PostForm)
	//fmt.Println(c)
	err := Save(c.FileName, c.Body)
	if err == nil {
		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride([]byte(c.Body))
		return
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(err.Error()))
		return
	}
	//Save(c.filename, c.body)
}
