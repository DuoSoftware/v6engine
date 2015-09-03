package unittesting

import (
	"duov6.com/objectstore/client"
	//"fmt"
)

func testClient() {
	//ytes, _ := client.Go("token", "com.duosoftware.customer", "account").GetOne().BySearching("supun").Ok()

	tmp := Account{}
	tmp.Id = "999"
	tmp.name = "SVD"
	tmp.address = "X"
	client.Go("token", "com.duosoftware.customer", "account").StoreObject().WithKeyField("Id").AndStoreOne(tmp).Ok()

	//if bytes != nil {
	//	fmt.Println(bytes)
	//}
}

type Account struct {
	Id      string
	name    string
	address string
}

//client.Go("4651687654b", "com.duosoftware.com", "account").StoreObject().WithKeyField("Id").AndStoreOne(obj).Ok()
