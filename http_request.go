package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
)

type HttpRequestType int64

var (
	HttpTLSTimeOut time.Duration = 10 // 10 seconds
	MyHTTPProxy    string        = "http://my-proxy-server.company.com:5050"
)

const (
	GET HttpRequestType = iota
	HEAD
	POST
	PUT
	PATCH
	DELETE
)

type UserCreds struct {
	User string
	Pass string
}

type Request struct {
	URL           string
	Payload       string
	RequestMethod HttpRequestType
	Headers       string
	ProxyURL      string
}

type RequestMap struct {
	URL           string
	Payload       interface{} // you are expected to set this as a string or map[string]interface{}
	RequestMethod HttpRequestType
	Headers       interface{} // you are expected to set this as a string or map[string]string
	ProxyURL      string
	Creds         UserCreds
}

func (r HttpRequestType) String() string {
	switch r {
	case GET:
		return "GET"
	case HEAD:
		return "HEAD"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case PATCH:
		return "PATCH"
	case DELETE:
		return "DELETE"
	}
	return "UNKNOWN"
}

type IHTTPRequest interface {
	HTTPRequest() (string, *http.Response, error)
}

type IHTTPRequestMap interface {
	HTTPRequest() (string, *http.Response, error)
}

func IsMap(data interface{}) bool {
	if data == nil {
		return false
	} else {
		return reflect.ValueOf(data).Kind() == reflect.Map
	}
}

func IsString(data interface{}) bool {
	if data == nil {
		return false
	} else {
		return reflect.ValueOf(data).Kind() == reflect.String
	}
}

func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func initRand() {
	rand.Seed(time.Now().UnixNano())
}

func (r RequestMap) HTTPRequest() (string, int, error) {
	log.Printf("URL : %v", r.URL)
	log.Printf("Payload : %v", r.Payload)
	log.Printf("RequestMethod : %v", r.RequestMethod)
	log.Printf("Headers : %v", r.Headers)

	responseString := ""
	responseHttpStatusCode := -1

	var headersMap map[string]interface{}
	var req *http.Request
	var err error
	var message string

	var tr *http.Transport

	payloadBytes := make([]byte, 0)
	payloadStr := ""

	headersBytes := make([]byte, 0)
	headersStr := ""

	// ---- r.Headers : should be a JSON string or map[string]string ------

	if IsString(r.Headers) {
		// if headers is string
		headersStr = r.Headers.(string)
		if !IsJSON(headersStr) {
			message = fmt.Sprintf("headers is not a valid JSON")
			log.Error(message)
			return responseString, responseHttpStatusCode, errors.New(message)
		}
		log.Printf("headers is a valid JSON STRING")
	} else {
		// if it is map[string]string or map[string]interface{}
		if IsMap(r.Headers) {
			log.Printf("headers is a valid map")
			headersBytes, err = json.Marshal(r.Headers)
			if err != nil {
				message = fmt.Sprintf("could not marshal json headers map to string : %v", err.Error())
				log.Error(message)
				return responseString, responseHttpStatusCode, errors.New(message)
			}
			headersStr = string(headersBytes)
			// log.Printf("headersStr : %v", headersStr)
		} else {
			// If is NOT a map
			log.Printf("headers is not a map")
			headersStr = ""
		}
	}

	// ---- r.Payload : should be a JSON string or map[string]interface{} ------

	if IsString(r.Payload) {
		// if it is a string
		payloadStr = r.Payload.(string)
		if !IsJSON(headersStr) {
			message = fmt.Sprintf("payload is not a valid JSON")
			log.Error(message)
			return responseString, responseHttpStatusCode, errors.New(message)
		}
		log.Printf("payload is a valid JSON string")
	} else {
		if IsMap(r.Payload) {
			// if it is map[string]string or map[string]interface{}
			log.Printf("payload is a valid map")
			payloadBytes, err = json.Marshal(r.Payload)
			if err != nil {
				message = fmt.Sprintf("could not marshal json payload map to string : %v", err.Error())
				log.Error(message)
				return responseString, responseHttpStatusCode, errors.New(message)
			}
			payloadStr = string(payloadBytes)
			log.Printf("payloadStr : %v", payloadStr)
		} else {
			// If is NOT a map
			log.Printf("payload is not a map")
			payloadStr = ""
		}
	}

	if r.ProxyURL == "" {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Dial: (&net.Dialer{
				Timeout:   0,
				KeepAlive: 0,
			}).Dial,
			TLSHandshakeTimeout: HttpTLSTimeOut * time.Second,
		}
	} else {
		proxyUrl, _ := url.Parse(MyHTTPProxy)
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Dial: (&net.Dialer{
				Timeout:   0,
				KeepAlive: 0,
			}).Dial,
			Proxy:               http.ProxyURL(proxyUrl),
			TLSHandshakeTimeout: HttpTLSTimeOut * time.Second,
		}
	}

	if payloadStr == "" {
		req, err = http.NewRequest(r.RequestMethod.String(), r.URL, nil)
		if err != nil {
			message = fmt.Sprintf("could not create a new http request : %v", err.Error())
			log.Error(message)
			return responseString, responseHttpStatusCode, errors.New(message)
		}
	} else {
		req, err = http.NewRequest(r.RequestMethod.String(), r.URL, bytes.NewBuffer([]byte(payloadStr)))
		if err != nil {
			message = fmt.Sprintf("could not create a new http request : %v", err.Error())
			log.Error(message)
			return responseString, responseHttpStatusCode, errors.New(message)
		}
	}

	if r.Creds.User != "" && r.Creds.Pass != "" {
		req.SetBasicAuth(r.Creds.User, r.Creds.Pass)
	}

	err = json.Unmarshal([]byte(headersStr), &headersMap)
	if err != nil {
		message = fmt.Sprintf("could not unmarshal headers json string : %v", err.Error())
		log.Error(message)
		return responseString, responseHttpStatusCode, errors.New(message)
	}

	if len(headersMap) != 0 {
		for key, value := range headersMap {
			valueStr, ok := value.(string)
			if !ok {
				continue
			}
			req.Header.Set(key, valueStr)
		}
	}

	req.Header.Set("Connection", "close")

	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		message = fmt.Sprintf("http client could not execute a http request : %v", err.Error())
		log.Error(message)
		return responseString, responseHttpStatusCode, errors.New(message)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		message = fmt.Sprintf("could not read http response body : %v", err.Error())
		log.Error(message)
		return responseString, responseHttpStatusCode, errors.New(message)
	}
	// convert body to string
	responseString = string(body)
	responseHttpStatusCode = resp.StatusCode
	return responseString, responseHttpStatusCode, nil
}
