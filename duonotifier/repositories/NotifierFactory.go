package repositories

import (
	"fmt"
)

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
