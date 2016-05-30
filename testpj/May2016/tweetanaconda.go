package main

import (
	"encoding/base64"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"io/ioutil"
	"net/url"
	"strconv"
)

func main() {
	PostImage()
}

func PostTweet() {
	anaconda.SetConsumerKey("wAkU0zuNYjMOqEmK1IkUoLe54")
	anaconda.SetConsumerSecret("7YjmfW6NbFV8wlpR4axwOhQQx5DHqMtcRy94wF6nrjCd2MlKNR")
	api := anaconda.NewTwitterApi("734985967442878466-lqRV7b2w4JZhRo2DKZh3JdEpdpzIFLc", "wF0js06M40K0YuLDOgAuCWZpMt4bIiEmpZ1moR4IqjaH8")

	result, err := api.PostTweet("TEST1", nil)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(result)
	}
}

func PostImage() {

	anaconda.SetConsumerKey("wAkU0zuNYjMOqEmK1IkUoLe54")
	anaconda.SetConsumerSecret("7YjmfW6NbFV8wlpR4axwOhQQx5DHqMtcRy94wF6nrjCd2MlKNR")
	api := anaconda.NewTwitterApi("734985967442878466-lqRV7b2w4JZhRo2DKZh3JdEpdpzIFLc", "wF0js06M40K0YuLDOgAuCWZpMt4bIiEmpZ1moR4IqjaH8")

	data, err := ioutil.ReadFile("smoothflow.jpg")
	if err != nil {
		fmt.Println(err.Error())
	}

	mediaResponse, err := api.UploadMedia(base64.StdEncoding.EncodeToString(data))
	if err != nil {
		fmt.Println(err.Error())
	}

	v := url.Values{}
	v.Set("media_ids", strconv.FormatInt(mediaResponse.MediaID, 10))

	result, err := api.PostTweet("Tweet with Image!", v)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(result)
	}
}
