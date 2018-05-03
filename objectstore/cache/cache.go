package cache

import (
	"duov6.com/objectstore/cache/repositories"
	"duov6.com/objectstore/messaging"
	"duov6.com/term"
	"encoding/json"
	"strings"
)

const (
	Data           = 0
	MetaData       = 1
	Transaction    = 3
	IncrementID    = 5
	RequestCounter = 6
	Log            = 8
)

func ResetSearchResults(request *messaging.ObjectRequest, database int) {
	_ = repositories.ResetSearchResultCache(request, database)
}

func DeleteOne(request *messaging.ObjectRequest, data map[string]interface{}, database int) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.RemoveOneRedis(request, data, database)
		if err != nil {
			term.Write("Error deleting one Object : "+err.Error(), term.Debug)
		}
	}
	return
}

func DeleteMany(request *messaging.ObjectRequest, data []map[string]interface{}, database int) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.RemoveManyRedis(request, data, database)
		if err != nil {
			term.Write("Error deleting many Objects : "+err.Error(), term.Debug)
		}
	}
	return
}

func Search(request *messaging.ObjectRequest, database int) (body []byte) {
	if CheckCacheAvailability(request) {
		body = repositories.GetSearch(request, database)
		if body != nil {
			bytesInString := string(body)
			bytesInString = strings.Replace(bytesInString, "\\u003e", ">", -1)
			bytesInString = strings.Replace(bytesInString, "\\u003c", "<", -1)
			bytesInString = strings.Replace(bytesInString, "u003e", ">", -1)
			bytesInString = strings.Replace(bytesInString, "u003c", "<", -1)
			body = []byte(bytesInString)
		}
	}
	return
}

func Query(request *messaging.ObjectRequest, database int) (body []byte) {
	if CheckCacheAvailability(request) {
		body = repositories.GetQuery(request, database)
		if body != nil {
			bytesInString := string(body)
			bytesInString = strings.Replace(bytesInString, "\\u003e", ">", -1)
			bytesInString = strings.Replace(bytesInString, "\\u003c", "<", -1)
			bytesInString = strings.Replace(bytesInString, "u003e", ">", -1)
			bytesInString = strings.Replace(bytesInString, "u003c", "<", -1)
			body = []byte(bytesInString)
		}
	}
	return
}

func GetByKey(request *messaging.ObjectRequest, database int) (body []byte) {
	if CheckCacheAvailability(request) {
		body = repositories.GetByKey(request, database)
		if body != nil {
			bytesInString := string(body)
			bytesInString = strings.Replace(bytesInString, "\\u003e", ">", -1)
			bytesInString = strings.Replace(bytesInString, "\\u003c", "<", -1)
			bytesInString = strings.Replace(bytesInString, "u003e", ">", -1)
			bytesInString = strings.Replace(bytesInString, "u003c", "<", -1)
			body = []byte(bytesInString)
		}
	}
	return
}

func StoreOne(request *messaging.ObjectRequest, data map[string]interface{}, database int) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.SetOneRedis(request, GetMapWithoutOsHeadersForStoreOne(data), database)
		if err != nil {
			term.Write("Error storing One Object to Cache : "+err.Error(), term.Debug)
		}
	}
	return
}

func GetMapWithoutOsHeadersForStoreOne(input map[string]interface{}) (output map[string]interface{}) {
	output = make(map[string]interface{})
	for key, value := range input {
		if key != "__osHeaders" {
			output[key] = value
		}
	}
	return
}

func GetMapWithoutOsHeadersForStoreMany(input []map[string]interface{}) (output []map[string]interface{}) {
	output = make([]map[string]interface{}, len(input))

	for x := 0; x < len(input); x++ {
		singleObject := make(map[string]interface{})
		for key, value := range input[x] {
			if key != "__osHeaders" {
				singleObject[key] = value
			}
		}
		output[x] = singleObject
	}
	return
}

func StoreKeyValue(request *messaging.ObjectRequest, key string, value string, database int) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.StoreKeyValue(request, key, value, database)
		if err != nil {
			term.Write("Error storing One Key Value to Cache : "+err.Error(), term.Debug)
		}
	}
	return
}

func GetKeyValue(request *messaging.ObjectRequest, key string, database int) (value []byte) {
	if CheckCacheAvailability(request) {
		value = repositories.GetKeyValue(request, key, database)
	}
	return
}

func ExistsKeyValue(request *messaging.ObjectRequest, key string, database int) (status bool) {
	if CheckCacheAvailability(request) {
		status = repositories.ExistsKeyValue(request, key, database)
	}
	return
}

func GetKeyListPattern(request *messaging.ObjectRequest, pattern string, database int) (value []string) {
	if CheckCacheAvailability(request) {
		value = repositories.GetKeyListPattern(request, pattern, database)
	}
	return
}

func DeleteKey(request *messaging.ObjectRequest, key string, database int) (status bool) {
	if CheckCacheAvailability(request) {
		status = repositories.DeleteKey(request, key, database)
	}
	return
}

func DeletePattern(request *messaging.ObjectRequest, pattern string, database int) (status bool) {
	if CheckCacheAvailability(request) {
		status = repositories.DeletePattern(request, pattern, database)
	}
	return
}

func StoreMany(request *messaging.ObjectRequest, data []map[string]interface{}, database int) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.SetManyRedis(request, GetMapWithoutOsHeadersForStoreMany(data), database)
		if err != nil {
			term.Write("Error storing Many Objects to Cache : "+err.Error(), term.Debug)
		}
	}
	return
}

func StoreResult(request *messaging.ObjectRequest, data interface{}, database int) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.SetResultRedis(request, data, database)
		if err != nil {
			term.Write("Error storing Get Result to Cache : "+err.Error(), term.Debug)
		}
	}
	return
}

func StoreQuery(request *messaging.ObjectRequest, data interface{}, database int) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.SetQueryRedis(request, data, database)
		if err != nil {
			term.Write("Error storing Query Result to Cache : "+err.Error(), term.Debug)
		}
	}
	return
}

func CheckCacheAvailability(request *messaging.ObjectRequest) (status bool) {
	status = true
	if request.Configuration.ServerConfiguration["REDIS"] == nil {
		//term.Write("Cache Config/Server Not Found!", term.Debug)
		status = false
	} else if !checkValidTenentClassKeywords(request) {
		status = false
	}

	return
}

func checkValidTenentClassKeywords(request *messaging.ObjectRequest) (status bool) {
	namespaces := [...]string{}
	keywords := [...]string{"UTC_TIMESTAMP()"}
	classes := [...]string{}
	//classes := [...]string{"b1", "c1", "SalesMaster", "RecurringLog", "ProfileMaster", "PaymentMaster", "InvoicedProducts"}
	status = true

	if request.Extras["IgnoreCacheRead"] != nil && request.Extras["IgnoreCacheRead"].(bool) == true {
		status = true
	} else {
		for _, namespace := range namespaces {
			if strings.ToLower(request.Controls.Namespace) == strings.ToLower(namespace) {
				term.Write("Invalid Namespace. Wouldn't be saved on Cache", term.Information)
				status = false
				return
			}
		}
		for _, class := range classes {
			if strings.ToLower(request.Controls.Class) == strings.ToLower(class) {
				term.Write("Invalid Class. Wouldn't be saved on Cache!", term.Information)
				status = false
				return
			}
		}
		for _, keyword := range keywords {
			byteArray, _ := json.Marshal(request)
			if strings.Contains(strings.ToLower(string(byteArray)), strings.ToLower(keyword)) {
				term.Write("Restrictive Keyword Found. Wouldn't be saved on Cache!", term.Information)
				status = false
				return
			}
		}
	}

	return
}

//Transaction Usage

func RPush(request *messaging.ObjectRequest, list string, value string, database int) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.RPush(request, list, value, database)
		if err != nil {
			term.Write("Error Rpushing to List : "+err.Error(), term.Debug)
		}
	}
	return
}

func LPush(request *messaging.ObjectRequest, list string, value string, database int) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.LPush(request, list, value, database)
		if err != nil {
			term.Write("Error Rpushing to List : "+err.Error(), term.Debug)
		}
	}
	return
}

func RPop(request *messaging.ObjectRequest, key string, database int) (result []byte, err error) {
	if CheckCacheAvailability(request) {
		result, err = repositories.RPop(request, key, database)
	}
	return
}

func LPop(request *messaging.ObjectRequest, key string, database int) (result []byte, err error) {
	if CheckCacheAvailability(request) {
		result, err = repositories.LPop(request, key, database)
	}
	return
}

func GetListLength(request *messaging.ObjectRequest, key string, database int) (length int64) {
	if CheckCacheAvailability(request) {
		length = repositories.GetListLength(request, key, database)
	}
	return
}

func FlushCache(request *messaging.ObjectRequest) {
	repositories.Flush(request)
}

func LRange(request *messaging.ObjectRequest, key string, database, start, end int) (result []string, err error) {
	if CheckCacheAvailability(request) {
		result, err = repositories.LRange(request, key, database, start, end)
	}
	return
}

func GetIncrValue(request *messaging.ObjectRequest, key string, database int) (val int64) {
	if CheckCacheAvailability(request) {
		val = repositories.GetIncrValue(request, key, database)
	}
	return
}
