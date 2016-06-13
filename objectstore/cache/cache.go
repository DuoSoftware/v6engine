package cache

import (
	"duov6.com/objectstore/cache/repositories"
	"duov6.com/objectstore/messaging"
	"duov6.com/term"
	"strings"
)

const (
	Data        = 0
	MetaData    = 1
	IncrementID = 5
	Transaction = 3
)

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
	}
	return
}

func Query(request *messaging.ObjectRequest, database int) (body []byte) {
	if CheckCacheAvailability(request) {
		body = repositories.GetQuery(request, database)
	}
	return
}

func GetByKey(request *messaging.ObjectRequest, database int) (body []byte) {
	if CheckCacheAvailability(request) {
		body = repositories.GetByKey(request, database)
	}
	return
}

func StoreOne(request *messaging.ObjectRequest, data map[string]interface{}, database int) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.SetOneRedis(request, data, database)
		if err != nil {
			term.Write("Error storing One Object to Cache : "+err.Error(), term.Debug)
		}
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
		err = repositories.SetManyRedis(request, data, database)
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
		term.Write("Cache Config/Server Not Found!", term.Debug)
		status = false
	} else if !checkValidTenentClass(request) {
		status = false
	}

	return
}

func checkValidTenentClass(request *messaging.ObjectRequest) (status bool) {
	namespaces := [...]string{}
	classes := [...]string{"domainclassattributes"}

	status = true

	for _, namespace := range namespaces {
		if strings.ToLower(request.Controls.Namespace) == strings.ToLower(namespace) {
			term.Write("Invalid Namespace. Wouldn't be saved on Cache", term.Debug)
			status = false
			return
		}
	}
	for _, class := range classes {
		if strings.ToLower(request.Controls.Class) == strings.ToLower(class) {
			term.Write("Invalid Class. Wouldn't be saved on Cache!", term.Debug)
			status = false
			return
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
