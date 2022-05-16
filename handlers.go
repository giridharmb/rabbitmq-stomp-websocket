package main

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
)

/*
HandlerBroadcast ...
*/
func HandlerBroadcast(w http.ResponseWriter, r *http.Request) {
	defer elapsed("__FUNC__: HandlerBroadcast")()
	errorMessage := ""
	returnResult := make(map[string]interface{})
	returnResult["error"] = ""
	returnResult["serverResponse"] = make([]interface{}, 0)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorMessage = fmt.Sprintf("Could not read request body !")
		log.Printf(errorMessage)
		returnResult["error"] = errorMessage
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&returnResult)
	}

	bodyStr := string(body)

	var myInterface interface{} = bodyStr
	_, ok := myInterface.(string)
	if !ok {
		errorMessage = fmt.Sprintf("HandlerBroadcast : String Assertion Failed !")
		log.Printf(errorMessage)
		returnResult["error"] = errorMessage
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&returnResult)
		return
	}

	parsedJSON, ok := gjson.Parse(bodyStr).Value().(map[string]interface{})
	if !ok {
		errorMessage = fmt.Sprintf("HandlerBroadcast: Could not convert to a map !")
		log.Printf(errorMessage)
		returnResult["error"] = errorMessage
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&returnResult)
		return
	}

	action, ok := parsedJSON["action"].(string)
	if !ok {
		errorMessage = fmt.Sprintf("HandlerBroadcast : 'action' is missing in the http request !")
		log.Printf(errorMessage)
		returnResult["error"] = errorMessage
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&returnResult)
		return
	}
	log.Printf("HandlerBroadcast : action : (%v)", action)

	queueName := ""
	queueName, ok = parsedJSON["queue_name"].(string)
	if ok {
		log.Printf("queueName : queueName : (%v)", queueName)
	}

	exchangeName := ""
	exchangeName, ok = parsedJSON["exchange_name"].(string)
	if ok {
		log.Printf("exchangeName : exchangeName : (%v)", exchangeName)
	}

	if queueName == "" && exchangeName == "" {
		errorMessage = fmt.Sprintf("HandlerBroadcast: only one of (exchange_name) or (queue_name) should have a valid value. Both cannot be set to some value. Both cannot be empty.")
		log.Printf(errorMessage)
		returnResult["error"] = errorMessage
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&returnResult)
		return
	}

	if queueName != "" && exchangeName != "" {
		errorMessage = fmt.Sprintf("HandlerBroadcast: only one of (exchange_name) or (queue_name) should have a valid value. Both cannot be set to some value. Both cannot be empty.")
		log.Printf(errorMessage)
		returnResult["error"] = errorMessage
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&returnResult)
		return
	}

	log.Printf("HandlerBroadcast : action : (%v)", action)

	messageAction(action, exchangeName, queueName, w, r)

}

func HandlerGetRabbitMQConnectionStatus(w http.ResponseWriter, r *http.Request) {
	defer elapsed("__FUNC__: HandlerGetData")()
	returnResult := make(map[string]interface{})
	returnResult["error"] = ""
	returnResult["serverResponse"] = make(map[string]interface{})
	status := IsAMQPConnectionClosed()
	returnResult["serverResponse"] = status
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&returnResult)
}
