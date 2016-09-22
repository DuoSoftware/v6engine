package cache

type CacheResponse struct {
	IsSuccess bool
	Body      []byte
}

type CacheRequest struct {
	Key    string
	Object interface{}
}
