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
	GetUsersShowById()
}

func GetUsersShowById() {
	anaconda.SetConsumerKey("Y4ZWzeC6dE4TU2SiEvdqOdGvI")
	anaconda.SetConsumerSecret("e4rjLOuEz5fQe3zUroVYLWQ2otVfNmyriOFcVU3wNghyeNEJBT")
	api := anaconda.NewTwitterApi("734985967442878466-1osCVq6ayMnbj6JnbN2ItlfsmTjnRBG", "LMfcek9DY5HuunbK2rAWVOehGBcUsIH660APwe78x2W3n")

	result, err := api.GetUsersShowById(734985967442878466, nil)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(result.Name)
		fmt.Println(result.IdStr)
		fmt.Println(result.ScreenName)
	}
}

func PostTweet() {
	anaconda.SetConsumerKey("Y4ZWzeC6dE4TU2SiEvdqOdGvI")
	anaconda.SetConsumerSecret("e4rjLOuEz5fQe3zUroVYLWQ2otVfNmyriOFcVU3wNghyeNEJBT")
	api := anaconda.NewTwitterApi("2698520514-nktdLnIebWzSLsOIxJSSrSYQzTEgvE6cc1sLggl", "egEM2AcLTzj7a6VZ5zPcorg88x11UjUnrZ2BtcyWOk5x8")

	result, err := api.PostTweet("HUEHUEHUE", nil)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(result)
	}
}

func GetHomeTimeline() {
	anaconda.SetConsumerKey("Y4ZWzeC6dE4TU2SiEvdqOdGvI")
	anaconda.SetConsumerSecret("e4rjLOuEz5fQe3zUroVYLWQ2otVfNmyriOFcVU3wNghyeNEJBT")
	api := anaconda.NewTwitterApi("2698520514-nktdLnIebWzSLsOIxJSSrSYQzTEgvE6cc1sLggl", "egEM2AcLTzj7a6VZ5zPcorg88x11UjUnrZ2BtcyWOk5x8")

	result, err := api.GetHomeTimeline(nil)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(result[10].Text)
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
