package main

import (
	"code.google.com/p/gorest"
	//"crypto/tls"
	"duov6.com/authlib"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	//"strings"
)

func invoke(res http.ResponseWriter, req *http.Request) {
	res.Header().Set(
		"Content-Type",
		"application/json",
	)
	var str string
	str = `<doctype html>
<html>
    <head>
        <title>Hello World</title>
    </head>
    <body>
        Hello World!
    </body>
</html>`

	term.Write("RequestURI "+req.RequestURI, term.Debug)
	term.Write("Header SecurityToken "+req.Header.Get("SecurityToken"), term.Error)
	term.Write("RemoteAddr "+req.RemoteAddr, term.Debug)
	term.Write("PostForm "+req.FormValue("method"), term.Debug)
	term.Write("PostForm "+req.FormValue("method"), term.Debug)
	term.Write("req.Method "+req.Method, term.Debug)
	//key)
	//

	a := authlib.Auth{}
	switch req.FormValue("method") {
	case "login":
		Auth := a.Login(req.FormValue("username"), req.FormValue("password"), req.FormValue("domain"))
		term.Write("PostForm "+req.FormValue("username"), term.Debug)
		term.Write("PostForm "+req.FormValue("password"), term.Debug)
		term.Write("PostForm "+req.FormValue("domain"), term.Debug)
		b, err := json.Marshal(Auth)
		if err == nil {
			//res.Header().Add("", value)
			io.WriteString(
				res,
				string(b),
			)
		}
		return

	default:
		io.WriteString(
			res,
			str,
		)

	}

}

func testChannel() {
	c1 := make(chan string)
	c2 := make(chan string)

	go func() {
		i1 := 0
		for {
			c1 <- "from 1 " + strconv.Itoa(i1)
			time.Sleep(time.Second * 1)
			i1++
		}

	}()
	go func() {
		i2 := 0
		for {
			c2 <- "from 2 " + strconv.Itoa(i2)
			time.Sleep(time.Second * 2)
			i2++
		}
	}()
	go func() {
		for {
			select {
			case msg1 := <-c1:
				fmt.Println(msg1)
			case msg2 := <-c2:
				fmt.Println(msg2)
			}
		}
	}()

	var input string
	fmt.Scanln(&input)
}

func main() {
	//testChannel()
	//http.Handle("/", gorest.Handle9())
	go webServer()
	go runRestFul()
	//term.Write(term.Read("What is your name"), term.Information)
	term.Write("Admintration Console running on :9000", term.Information)
	term.Write("https RestFul Service running on :3048", term.Information)
	s := ""
	num, err := fmt.Scanln(&s)
	fmt.Println(s)
	fmt.Println(num)
	fmt.Println(err)
	term.StartCommandLine()

}

func status() {
	term.Write("Status is running", term.Information)
}

func webServer() {
	http.Handle(
		"/",
		http.StripPrefix(
			"/",
			http.FileServer(http.Dir("html")),
		),
	)
	http.ListenAndServe(":9000", nil)
}

func runRestFul() {
	gorest.RegisterService(new(authlib.Auth))
	err := http.ListenAndServeTLS(":3048", "apache.crt", "apache.key", gorest.Handle())
	if err != nil {
		term.Write(err.Error(), term.Error)
		return
	}
}
