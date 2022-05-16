package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func messageAction(action string, exchangeName string, queueName string, w http.ResponseWriter, r *http.Request) {

	errorMessage := ""
	returnResult := make(map[string]interface{})
	returnResult["error"] = ""
	returnResult["serverResponse"] = make([]interface{}, 0)
	switch action {

	case "broadcast_all":
		log.Printf(">> (MQ) : OPENING CONNECTION...")

		var queueAndExchange = QueueAndExchange{
			QueueName:    "",
			ExchangeName: exchangeName,
		}

		mq := ConnectionDetails{
			MQUser:           MQUSER,
			MQPass:           MQPASS,
			MQHost:           MQHOST,
			QueueAndExchange: queueAndExchange,
		}

		rabbitConnect, queueAndExchange, err := mq.OpenConnection()
		if err != nil {
			log.Printf(">> (MQ) : error : %v", err.Error())
			errorMessage = fmt.Sprintf("HandlerBroadcast : could not create MQ connection : %v", err.Error())
			log.Printf(errorMessage)
			returnResult["error"] = errorMessage
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(&returnResult)
			return
		}
		pushMessagesGeneric(queueAndExchange, rabbitConnect)

		/* **************************************************************************************** */
	case "broadcast_session":
		if queueName == "" {
			errorMessage = fmt.Sprintf("HandlerBroadcast : 'queue_name' cannot be empty string !")
			log.Printf(errorMessage)
			returnResult["error"] = errorMessage
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(&returnResult)
			return
		}
		log.Printf(">> (MQ) : OPENING CONNECTION...")

		var queueAndExchange = QueueAndExchange{
			QueueName:    queueName,
			ExchangeName: "",
		}

		mq := ConnectionDetails{
			MQUser:           MQUSER,
			MQPass:           MQPASS,
			MQHost:           MQHOST,
			QueueAndExchange: queueAndExchange,
		}

		rabbitConnect, queueAndExchange, err := mq.OpenConnection()
		if err != nil {
			log.Printf(">> (MQ) : error : %v", err.Error())
			errorMessage = fmt.Sprintf("HandlerBroadcast : could not create MQ connection : %v", err.Error())
			log.Printf(errorMessage)
			returnResult["error"] = errorMessage
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(&returnResult)
			return
		}

		pushMessagesGeneric(queueAndExchange, rabbitConnect)

		/* **************************************************************************************** */
	default:
		errorMessage = fmt.Sprintf("HandlerBroadcast : please provide valid 'action' in the payload !")
		log.Printf(errorMessage)
		returnResult["error"] = errorMessage
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&returnResult)
		return
		/* **************************************************************************************** */
	}

	returnResult["serverResponse"] = "successfully_broadcasted"
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&returnResult)

}
