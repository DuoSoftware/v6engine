package repositories

import (
	"crypto/tls"
	"duov6.com/duonotifier/messaging"
	"duov6.com/objectstore/client"
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

type EmailNotifier struct {
}

type Emailtemplate struct {
	Id         string
	Subject    string
	Body       string
	Signature  string
	Parameters map[int]string
}

func (notifier EmailNotifier) GetNotifierName() string {
	return "EmailNotifier"
}

func (notifier EmailNotifier) ExecuteNotifier(request *messaging.NotifierRequest) messaging.NotifierResponse {
	var response = messaging.NotifierResponse{}
	securityToken := request.Controls.SecurityToken
	namespace := request.Controls.Namespace
	class := request.Controls.Class
	templateId := request.Parameters["templateId"].(string)
	inputParameters := request.Parameters["parameters"].(map[string]string)
	var reciever string
	reciever = ""
	var recievers map[int]string

	if request.Parameters["reciever"] != nil {
		reciever = request.Parameters["reciever"].(string)
	}

	if request.Parameters["cc_mail"] != nil {
		recievers = request.Parameters["cc_mail"].(map[int]string)
	}

	if recievers == nil && (reciever == "" || reciever == " ") {
		response.Message = "No Recievers included"
		response.IsSuccess = false
	} else {

		response.IsSuccess = send(request, securityToken, namespace, class, templateId, inputParameters, reciever, recievers)

		if response.IsSuccess {
			response.Message = "All Emails sent Successfully!"
		} else {
			response.Message = "Email sending Failure. Check parameters and connection!"
		}
	}

	return response
}

func send(request *messaging.NotifierRequest, securityToken string, domain string, class string, templateId string, inputParams map[string]string, recieverEmail string, multiplerevieverEmails map[int]string) (isSuccess bool) {
	isSuccess = true
	var recievedEmailData Emailtemplate
	recievedEmailData = getEmailData(securityToken, domain, class, templateId)
	recievedEmailData = convert(recievedEmailData, inputParams)

	isGlobalSendingSuccess := true

	if recieverEmail != "" && recieverEmail != " " {
		isSuccess = sendmail(request, recieverEmail, recievedEmailData.Subject, (recievedEmailData.Body + "\r\n \r\n" + recievedEmailData.Signature))

		if isSuccess {
			isGlobalSendingSuccess = true
		} else {
			isGlobalSendingSuccess = false
		}
	}

	if multiplerevieverEmails != nil {
		for _, reciever := range multiplerevieverEmails {
			isSuccess = sendmail(request, reciever, recievedEmailData.Subject, (recievedEmailData.Body + "\r\n \r\n" + recievedEmailData.Signature))

			if !isSuccess {
				isGlobalSendingSuccess = false
			}
		}
	}

	isSuccess = isGlobalSendingSuccess
	return
}

func getEmailData(securityToken string, domain string, class string, templateId string) (email Emailtemplate) {
	email = Emailtemplate{}

	bytes, _ := client.Go(securityToken, domain, class).GetOne().ByUniqueKey(templateId).Ok()

	email = Emailtemplate{}

	var array map[string]interface{}
	array = make(map[string]interface{})
	_ = json.Unmarshal(bytes, &array)

	for key, value := range array {
		if key != "__osHeaders" {
			if key == "Id" {
				email.Id = value.(string)
				continue
			} else if key == "Subject" {
				email.Subject = value.(string)
				continue
			} else if key == "Body" {
				email.Body = value.(string)
				continue
			} else if key == "Signature" {
				email.Signature = value.(string)
				continue
			} else if key == "Parameters" {
				email.Parameters = getParameterMap(value.(string))
			}
		}
	}

	return email
}

func convert(email Emailtemplate, substitue map[string]string) (retEmail Emailtemplate) {
	retEmail = Emailtemplate{}

	for key, value := range substitue {
		email.Subject = strings.Replace(email.Subject, ("@" + key + "@"), value, -1)
		email.Body = strings.Replace(email.Body, ("@" + key + "@"), value, -1)
		email.Signature = strings.Replace(email.Signature, ("@" + key + "@"), value, -1)
	}

	retEmail.Id = email.Id
	retEmail.Subject = email.Subject
	retEmail.Body = email.Body
	retEmail.Signature = email.Signature
	retEmail.Parameters = email.Parameters

	return retEmail
}

func getParameterMap(paratermeters string) (returnMap map[int]string) {
	returnMap = make(map[int]string)

	tokens := strings.Split(paratermeters, ",")

	for key, value := range tokens {
		returnMap[key] = value
	}

	return returnMap
}

func sendmail(request *messaging.NotifierRequest, receiver string, subj string, body string) (isSuccessful bool) {
	isSuccessful = false
	password := request.Configuration.NotifyMethodsConfig["EMAIL"]["Password"]
	server := request.Configuration.NotifyMethodsConfig["EMAIL"]["Server"]
	username := request.Configuration.NotifyMethodsConfig["EMAIL"]["Email"]

	from := mail.Address{"", username}
	to := mail.Address{"", receiver}
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj
	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body
	// Connect to the SMTP Server
	servername := server
	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", username, password, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	fmt.Print(tlsconfig.Certificates)

	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		//fmt.Print(err.Error())
		fmt.Println("Cannot Create TCP Link to : " + servername)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		//fmt.Print(err.Error())
		fmt.Println("Cannot Create New Client to : " + host)
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		//fmt.Print(err.Error())
		fmt.Println("Error Connecting to Authentication")
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		//fmt.Print(err.Error())
		fmt.Println("Error emailing..")
	}

	if err = c.Rcpt(to.Address); err != nil {
		//fmt.Print(err.Error())
		fmt.Println("Error Contacting address")
	}

	// Data
	w, err := c.Data()
	if err != nil {
		//fmt.Print(err.Error())
		fmt.Println("Error parsing Data")
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		//fmt.Print(err.Error())
		fmt.Println("Error writing Data")
	}

	err = w.Close()
	if err != nil {
		//fmt.Print(err.Error())
		fmt.Println("\nMail sending FAILED to address : " + receiver)
	} else {
		//fmt.Println("\nMail sent sucessfully to address : " + receiver)
		isSuccessful = true
	}

	c.Quit()
	return
}
