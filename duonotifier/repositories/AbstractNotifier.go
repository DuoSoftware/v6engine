package repositories

import (
	"duov6.com/duonotifier/messaging"
)

type AbstractNotifier interface {
	GetNotifierName() string
	ExecuteNotifier(request *messaging.NotifierRequest) messaging.NotifierResponse
}
