package common

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

func HTTP_POST(url string, headers map[string]string, JSON_DATA []byte, isURLParams bool) (err error, body []byte) {

	if isURLParams {
		url = strings.TrimSuffix(url, "/")
		if len(headers) > 0 {
			if strings.Contains(url, "?") {
				url += "&"
			} else {
				url += "?"
			}
		}

		for headerName, headerValue := range headers {
			url += headerName + "=" + headerValue + "&"
		}

		url = strings.TrimSuffix(url, "&")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(JSON_DATA))

	if !isURLParams {
		for headerName, headerValue := range headers {
			req.Header.Set(headerName, headerValue)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = errors.New("Connection Failed!")
	} else {
		defer resp.Body.Close()
		body, _ = ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 && resp.StatusCode != 204 && resp.StatusCode != 201 {
			err = errors.New(string(body))
		}
	}
	//defer resp.Body.Close()
	return
}

func HTTP_PUT(url string, headers map[string]string, JSON_DATA []byte, isURLParams bool) (err error, body []byte) {

	if isURLParams {
		url = strings.TrimSuffix(url, "/")
		if len(headers) > 0 {
			if strings.Contains(url, "?") {
				url += "&"
			} else {
				url += "?"
			}
		}

		for headerName, headerValue := range headers {
			url += headerName + "=" + headerValue + "&"
		}

		url = strings.TrimSuffix(url, "&")
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(JSON_DATA))

	if !isURLParams {
		for headerName, headerValue := range headers {
			req.Header.Set(headerName, headerValue)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = errors.New("Connection Failed!")
	} else {
		defer resp.Body.Close()
		body, _ = ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 && resp.StatusCode != 204 {
			err = errors.New(string(body))
		}
	}
	//defer resp.Body.Close()
	return
}

func HTTP_DELETE(url string, headers map[string]string, JSON_DATA []byte, isURLParams bool) (err error, body []byte) {

	if isURLParams {
		url = strings.TrimSuffix(url, "/")
		if len(headers) > 0 {
			if strings.Contains(url, "?") {
				url += "&"
			} else {
				url += "?"
			}
		}

		for headerName, headerValue := range headers {
			url += headerName + "=" + headerValue + "&"
		}

		url = strings.TrimSuffix(url, "&")
	}

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(JSON_DATA))

	if !isURLParams {
		for headerName, headerValue := range headers {
			req.Header.Set(headerName, headerValue)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = errors.New("Connection Failed!")
	} else {
		defer resp.Body.Close()
		body, _ = ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 && resp.StatusCode != 204 {
			err = errors.New(string(body))
		}
	}
	// defer resp.Body.Close()
	return
}

func HTTP_GET(url string, headers map[string]string, isURLParams bool) (err error, body []byte) {

	if isURLParams {
		url = strings.TrimSuffix(url, "/")
		if len(headers) > 0 {
			if strings.Contains(url, "?") {
				url += "&"
			} else {
				url += "?"
			}
		}

		for headerName, headerValue := range headers {
			url += headerName + "=" + headerValue + "&"
		}

		url = strings.TrimSuffix(url, "&")
	}

	req, err := http.NewRequest("GET", url, nil)

	if !isURLParams {
		for headerName, headerValue := range headers {
			req.Header.Set(headerName, headerValue)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = errors.New("Connection Failed!")
	} else {
		defer resp.Body.Close()
		body, _ = ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 && resp.StatusCode != 204 {
			err = errors.New(string(body))
		}
	}
	// defer resp.Body.Close()
	return
}
