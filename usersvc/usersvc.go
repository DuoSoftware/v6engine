package main

import (
	//"bytes"
	"code.google.com/p/gorest"
	"crypto/rand"
	"crypto/tls"
	//"duov6.com/applib"
	"duov6.com/cebadapter"
	"duov6.com/common"
	"encoding/json"
	"fmt"
	//"io/ioutil"
	//"log"
	//"duov6.com/config"
	"duov6.com/objectstore/client"
	"duov6.com/term"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
)

type User struct {
	UserID          string
	EmailAddress    string
	Name            string
	Password        string
	ConfirmPassword string
	Active          bool
}

type Registation struct {
	UserID          string
	EmailAddress    string
	Password        string
	Name            string
	ConfirmPassword string
}
type ResetEmail struct {
	ResetEmail string
}
/*
type AuthHandler struct {
	//Config AuthConfig
}*/
type Password struct {
	EmailAddress string
	Password     string
}
type Login struct {
	EmailAddress string
	Password     string
}
type ActivationEmail struct {
	EmailAddress string
	Token        string
}

type AuthConfig struct {
	Cirtifcate    string
	PrivateKey    string
	Https_Enabled bool
	StoreID       string
	Smtpserver    string
	Smtpusername  string
	Smtppassword  string
	UserName      string
	Password      string
}
/*
func newAuthHandler() *AuthHandler {
	authhld := new(AuthHandler)
	//authhld.Config = GetConfig()
	return authhld
}*/

var Config AuthConfig

//Service Definition
type RegistationService struct {
	gorest.RestService
	//gorest.RestService `root:"/tutorial/"`
	userRegistation gorest.EndPoint `method:"POST" path:"/UserRegistation/" postdata:"Registation"`
	userActivation  gorest.EndPoint `method:"GET" path:"/UserActivation/{token:string}" output:"string"`
	login           gorest.EndPoint `method:"POST" path:"/Login/" postdata:"Login"`
	//NOT COMPLETER 100% yet Validation and change to Authlib methods
	/*resetPassword gorest.EndPoint `method:"POST" path:"/ResetPassword/" postdata:"ResetEmail"`
	passwordSet   gorest.EndPoint `method:"GET" path:"/PasswordSet/{token:string}" output:"string"`
	passwordSave  gorest.EndPoint `method:"POST" path:"/PasswordSave/" postdata:"Password"`*/
}

func main() {
	cebadapter.Attach("Registration", func(s bool) {
		cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
			fmt.Println("Store Configuration Successfully Loaded...")

			agent := cebadapter.GetAgent()

			agent.Client.OnEvent("globalConfigChanged.StoreConfig", func(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
				cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
					fmt.Println("Store Configuration Successfully Updated...")
				})
			})
		})
		fmt.Println("Successfully registered in CEB")
	})

	gorest.RegisterService(new(RegistationService)) //Register our service
	http.Handle("/", gorest.Handle())
	argument := os.Args[1]
	fmt.Println(argument)
	http.ListenAndServe(":"+argument, nil)
}

//Register new user
//POST
//POST DATA sample {"EmailAddress":"pamidu@duosoftware.com","Name":"Pamidu","Password":"admin","ConfirmPassword":"admin"}
func (serv RegistationService) UserRegistation(r Registation) {
	var user User
	user.Active = false
	user.ConfirmPassword = r.ConfirmPassword
	user.EmailAddress = r.EmailAddress
	user.Name = r.Name
	user.Password = r.Password
	//fmt.Println("SAVE USER\n\n\n")
	//Save user Method
	fmt.Println("1")
	res := SaveUser(user)
	fmt.Println("2")
	fmt.Println(res)
	serv.ResponseBuilder().SetResponseCode(200).Write([]byte("done..."))

}

//Save user using Authlib
func SaveUser(u User) string {
	term.Write("SaveUser saving user  "+u.Name, term.Debug)
	respond := ""
	token := randToken()
	fmt.Println("3")
	fmt.Println(">>>>>>>>>>>>>>>")
	fmt.Println(u.EmailAddress)
	fmt.Println("<<<<<<<<<<<<<<<")
	//bytes, err := client.Go("ignore", "com.duosoftware.auth", "users").GetOne().ByQuerying("EmailAddress :" + "prasadacicts@gmail.com").Ok()
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "users").GetOne().ByUniqueKey(u.EmailAddress).Ok()

	fmt.Println("{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{")
	fmt.Println(string(bytes))
	fmt.Println("}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}")
	fmt.Println("4")
	if err == "" {
		var uList []User
		uList = make([]User, 0)
		err := json.Unmarshal(bytes, &uList)
		if err == nil || bytes == nil {
			fmt.Println("5")
			//new user

			fmt.Println(len(uList), "LLLLLLLLLLLLLLLLLLLL")
			if len(uList) == 0 {
				u.UserID = common.GetGUID()
				term.Write("SaveUser saving user"+u.Name+" New User "+u.UserID, term.Debug)
				client.Go("ignore", "com.duosoftware.auth", "users").StoreObject().WithKeyField("EmailAddress").AndStoreOne(u).Ok()
				respond = "true"
				fmt.Println("6")
				//save Activation mail details
				//EmailAddress and Token
				//EmailAddress KeyProperty
				var Activ ActivationEmail
				Activ.EmailAddress = u.EmailAddress
				Activ.Token = token
				client.Go("ignore", "com.duosoftware.com", "activation").StoreObject().WithKeyField("EmailAddress").AndStoreOne(Activ).Ok()
				fmt.Println("7")
				Email(u.EmailAddress, token, "Activation")

			} else if len(uList) == 1 {
				//Alredy in  Registerd user
				//term.Write("User Already Registerd  #"+err.Error(), term.Error)
				fmt.Println("User Already Registerd")

			}
		} else {
			fmt.Println("ERRRRRR")
			//term.Write("SaveUser saving user store Error #"+err.Error(), term.Error)
			respond = "false"

		}
	} else {
		//term.Write("SaveUser saving user fetech Error #"+err, term.Error)
		fmt.Println("errrrr")
		respond = "false"
	}
	fmt.Println("8")
	u.Password = "*****"
	u.ConfirmPassword = "*****"
	return respond
}

//Activate user account using invitation mail send with token
//GET
//Url  /UserActivation/sdfsdfwer23rsdff
//if user activation success method will return Success
func (serv RegistationService) UserActivation(token string) string {
	respond := ""
	//check user from db
	bytes, err := client.Go("ignore", "com.duosoftware.com", "activation").GetOne().BySearching(token).Ok()
	if err == "" {
		var uList []User
		err := json.Unmarshal(bytes, &uList)
		if err == nil || bytes == nil {
			//new user
			if len(uList) == 0 {

				term.Write("User Not Found", term.Debug)

			} else {
				var u User
				u.UserID = uList[0].UserID
				u.Password = uList[0].Password
				u.Active = true
				u.ConfirmPassword = uList[0].Password
				u.Name = uList[0].Name
				u.EmailAddress = uList[0].EmailAddress

				//Change activation status to true and save
				term.Write("Activate User  "+u.Name+" Update User "+u.UserID, term.Debug)
				client.Go("ignore", "com.duosoftware.auth", "users").StoreObject().WithKeyField("EmailAddress").AndStoreOne(u).Ok()
				respond = "true"
				var Activ ActivationEmail
				Activ.EmailAddress = u.EmailAddress
				//set token empty and save
				Activ.Token = ""
				client.Go("ignore", "com.duosoftware.com", "Activation").StoreObject().WithKeyField("EmailAddress").AndStoreOne(Activ).Ok()

				Email(u.EmailAddress, Activ.Token, "Activated")
				respond = "Success"
			}
		}

	} else {
		term.Write("Activation Fail ", term.Debug)

	}

	return respond

}

//User login
func (serv RegistationService) Login(l Login) {
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "users").GetOne().ByUniqueKey(l.EmailAddress).Ok()
	fmt.Println(l.EmailAddress, l.Password)
	fmt.Println("19")
	if err == "" {
		fmt.Println("20")
		if bytes != nil {
			fmt.Println("21")
			newUser := User{}
			//uList = make([]User, 0)
			err := json.Unmarshal(bytes, &newUser)
			fmt.Println("22")
			fmt.Println("<<<<<<<<<<<", newUser)

			if err == nil {
				fmt.Println("23")
				if newUser.Password == l.Password && newUser.EmailAddress == l.EmailAddress {
					fmt.Println("24")
					serv.ResponseBuilder().SetResponseCode(200).Write([]byte(newUser.Name))
					//term.Write("password incorrect", term.Error)
				} else {
					fmt.Println("25")
					serv.ResponseBuilder().SetResponseCode(201).Write([]byte("Password Wrong "))
				}
			} else {
				fmt.Println("26")
				if err != nil {
					fmt.Println("27")
					term.Write("Login  user Error "+err.Error(), term.Error)
				}
			}
		}
	} else {
		fmt.Println("28")
		term.Write("Login  user  Error "+err, term.Error)
		serv.ResponseBuilder().SetResponseCode(201).Write([]byte(err))
	}

}

//send Activation ,Passowrd Reset request and password change mail
//message contating not set properly
func Email(receiver, token string, emailtype string) string {
	res := "FALSE"
	from := mail.Address{"", "pamidu@duosoftware.com"}
	to := mail.Address{"", receiver}
	subj := ""
	body := ""

	fmt.Println("9")

	if emailtype == "Activation" {
		fmt.Println("10")
		subj = "DuoWorld Activation Requierd"
		body = "<html><head> <title></title> <link rel=\"stylesheet\" href=\"https://maxcdn.bootstrapcdn.com/bootstrap/3.3.4/css/bootstrap.min.css\"> </head><body><section style=\"position: relative;padding: 60px 0 60px 0;width: 869px;height: 493px;background: rgb(40, 70, 102) url('http://i58.tinypic.com/2cpp2bq.jpg') no-repeat center center; background-size: cover;color: #fff;\"><div class=\"row hero-conten\"> <div class=\"col-md-12 text-center\"><img src=\"http://i57.tinypic.com/2qbx3c7.png\" alt=\"DuoWorld Logo\" width=\"50%\" height=\"30%\"></div> <div class=\"col-md-12 text-center\"> <p><h2>Just one more step...</h2></p> <p><h4>Click the Activate button below to activate your DuoWorld account.</h4></p><br/>  <button class=\"btn\" style=\"width: 100px;height: 30px;font-size: 27px;background-color: aquamarine;\"> <a href=\"http://duoworld.sossgrid.com:1000/UserActivation/" + token + "\">Activate</a></button></div></div></section></body></html>"

	} else if emailtype == "Activated" {
		subj = "This is Activated "
		body = "Click To Reset Password.\n With two line"

	} else if emailtype == "PasswordReset" {
		subj = "This is Password Reset Request "
		body = "Click To Reset Password.\n With two lines\n http://duoworld.sossgrid.com:1000/PasswordSet/" + token

	} else if emailtype == "PasswordSetSuccess" {
		subj = "This is Password set success"
		body = "password set success"
	}
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj
	headers["Content-type"] = "text/html"
	fmt.Println("11")
	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body
	// Connect to the SMTP Server
	servername := "173.194.65.108:465"
	host, _, _ := net.SplitHostPort(servername)
	auth := smtp.PlainAuth("", "pamidu@duosoftware.com", "DuoS@123", host)

	// TLS config
	fmt.Print("12")
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	//fmt.Print("8")
	fmt.Println("13")
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("14")
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		fmt.Println(err.Error())
	}
	// Auth
	fmt.Println("15")
	if err = c.Auth(auth); err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("16")
	// To && From
	if err = c.Mail(from.Address); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("17")
	fmt.Println(to.Address)
	if err = c.Rcpt(to.Address); err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("18")
	// Data
	w, err := c.Data()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("19")
	_, err = w.Write([]byte(message))
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("20")
	err = w.Close()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("\nMail sent sucessfully....")
		res = "TRUE"
	}
	fmt.Println("21")
	c.Quit()
	return res

}

//genarate random token
func randToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
