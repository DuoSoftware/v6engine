package drivers

import (
	"duov6.com/objectstore/messaging"
)

func getDomainClassAttributesKey(request *messaging.ObjectRequest) (key string) {
	key = request.Controls.Namespace + ".domainClassAttributes." + request.Controls.Class
	return
}
