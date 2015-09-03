package repositories

import (
	"duov6.com/duonotifier/messaging"
)

type SMSNotifier struct {
}

func (notifier SMSNotifier) GetNotifierName() string {
	return "SMSNotifier"
}

func (notifier SMSNotifier) ExecuteNotifier(request *messaging.NotifierRequest) messaging.NotifierResponse {
	var temp = messaging.NotifierResponse{}
	temp.IsSuccess = false
	temp.Message = "Not Implemented in SMSNotifier."
	return temp
}
