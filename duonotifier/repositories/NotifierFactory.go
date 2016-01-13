package repositories

import (
	"duov6.com/duonotifier/messaging"
	"fmt"
)

func Execute(Request *messaging.NotifierRequest) messaging.NotifierResponse {
	var response messaging.NotifierResponse
	abstractRepository := Create(Request.NotifyMethod)
	fmt.Println("Executing Abstract Repository : " + abstractRepository.GetNotifierName())
	response = abstractRepository.ExecuteNotifier(Request)
	return response
}

func Create(code string) AbstractNotifier {

	fmt.Println("Excuting AbstractNotifier : " + code)

	var notifier AbstractNotifier
	switch code {
	case "EMAIL":
		notifier = EmailNotifier{}
	case "SMS":
		notifier = SMSNotifier{}
	}
	return notifier
}
