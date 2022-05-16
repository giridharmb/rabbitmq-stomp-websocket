package main

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"time"
)

//var RabbitMQConnectionURL string = ""
//var RabbitMQConnection *amqp.Connection
//var RabbitMQChannel *amqp.Channel

type BroadcastType int64

const (
	SendForSession BroadcastType = iota
	SendForAll
)

type ConnectionDetails struct {
	MQUser           string
	MQPass           string
	MQHost           string
	QueueAndExchange QueueAndExchange
}

type QueueAndExchange struct {
	QueueName    string
	ExchangeName string
}

type ConnectionChannelQueue struct {
	MQConnection *amqp.Connection
	MQChannel    *amqp.Channel
	MQQueue      amqp.Queue
	Broadcast    BroadcastType
}

type IGenericMQConnectionBroadcast interface {
	OpenConnection(mqConn ConnectionDetails) (*ConnectionChannelQueue, QueueAndExchange, error)
}

type IGenericRabbitConnChanQueue interface {
	Close(rabbitConnectionChannelAndQueue ConnectionChannelQueue)
}

func CreateFanoutExchange(exchange string, vhost string) error {
	log.Printf("CreateFanoutExchange : exchange : (%v)", exchange)
	msg := ""
	if exchange == "" {
		msg = fmt.Sprintf("CreateFanoutExchange() : exchange cannot be EMPTY !")
		return errors.New(msg)
	}
	headers := fmt.Sprintf(`{ 
		"Content-Type" : "application/json",
		"Accept" : "application/json"
	}`)

	payload := fmt.Sprintf(`{"type":"fanout","durable":true}`)

	creds := UserCreds{
		User: MQUSER,
		Pass: MQPASS,
	}

	request := RequestMap{
		URL:           "http://" + MQHOST + ":15672/api/exchanges/" + vhost + "/" + exchange,
		Payload:       payload,
		RequestMethod: PUT,
		Headers:       headers,
		ProxyURL:      "",
		Creds:         creds,
	}

	response, responseStatusCode, err := request.HTTPRequest()
	if err != nil {
		log.Printf("error : could not make http request : %v", err.Error())
		return err
	}
	log.Printf("CreateFanoutExchange() : response : %v", response)
	log.Printf("CreateFanoutExchange() : responseStatusCode : %v", responseStatusCode)
	return nil
}

func SetVHostPermissions(userName string, vhost string) error {
	log.Printf("SetVHostPermissions : exchange : (%v)", exchange)
	msg := ""
	if userName == "" || vhost == "" {
		msg = fmt.Sprintf("SetVHostPermissions() : (userName / vhost) cannot be EMPTY !")
		return errors.New(msg)
	}
	headers := fmt.Sprintf(`{ 
		"Content-Type" : "application/json",
		"Accept" : "application/json"
	}`)

	payload := fmt.Sprintf(`{"configure":".*","write":".*","read":".*"}`)

	creds := UserCreds{
		User: MQUSER,
		Pass: MQPASS,
	}

	request := RequestMap{
		URL:           "http://" + MQHOST + ":15672/api/permissions/" + vhost + "/" + userName,
		Payload:       payload,
		RequestMethod: PUT,
		Headers:       headers,
		ProxyURL:      "",
		Creds:         creds,
	}

	response, responseStatusCode, err := request.HTTPRequest()
	if err != nil {
		log.Printf("error : could not make http request : %v", err.Error())
		return err
	}
	log.Printf("SetVHostPermissions() : response : %v", response)
	log.Printf("SetVHostPermissions() : responseStatusCode : %v", responseStatusCode)
	return nil
}

func CreateVHost(vhost string) error {

	log.Printf("CreateVHost : vhost : (%v)", vhost)

	msg := ""
	if vhost == "" {
		msg = fmt.Sprintf("CreateVHost() : vhost cannot be EMPTY !")
		return errors.New(msg)
	}
	headers := fmt.Sprintf(`{ 
		"Content-Type" : "application/json",
		"Accept" : "application/json"
	}`)

	creds := UserCreds{
		User: MQUSER,
		Pass: MQPASS,
	}

	request := RequestMap{
		URL:           "http://" + MQHOST + ":15672/api/vhosts/" + vhost,
		Payload:       nil,
		RequestMethod: PUT,
		Headers:       headers,
		ProxyURL:      "",
		Creds:         creds,
	}

	response, responseStatusCode, err := request.HTTPRequest()
	if err != nil {
		log.Printf("error : could not make http request : %v", err.Error())
		return err
	}
	log.Printf("CreateVHost() : response : %v", response)
	log.Printf("CreateVHost() : responseStatusCode : %v", responseStatusCode)
	return nil
}

func IsAMQPConnectionClosed() bool {
	var mqURL string
	var err error
	var mqConnection *amqp.Connection
	var connectionClosed bool
	mqURL = fmt.Sprintf("amqp://%v:%v@%v:5672/%v", MQUSER, MQPASS, MQHOST, vhost)
	mqConnection, err = amqp.DialConfig(mqURL, amqp.Config{Heartbeat: 5 * time.Second})
	if err != nil {
		log.Printf("error : CheckAMQPConnection() : %v", err.Error())
		return false
	} else {
		connectionClosed = mqConnection.IsClosed()
		_ = mqConnection.Close()
		return connectionClosed
	}
}

func (connDetails ConnectionDetails) OpenConnection() (ConnectionChannelQueue, QueueAndExchange, error) {
	defer elapsed("__FUNC__: OpenConnection")()
	log.Printf(">> (MQ) : OPENING CONNECTION...")
	msg := ""
	var err error
	var mqURL string
	var mqChannel *amqp.Channel
	var mqConnection *amqp.Connection
	var mqQueue amqp.Queue

	var queueAndExchange = QueueAndExchange{
		QueueName:    "",
		ExchangeName: "",
	}

	var rabbitConnectionChannelAndQueue = ConnectionChannelQueue{
		MQConnection: nil,
		MQChannel:    nil,
		MQQueue:      amqp.Queue{},
		Broadcast:    SendForAll,
	}

	mqURL = fmt.Sprintf("amqp://%v:%v@%v:5672/%v", connDetails.MQUser, connDetails.MQPass, connDetails.MQHost, vhost)
	mqConnection, err = amqp.DialConfig(mqURL, amqp.Config{Heartbeat: 5 * time.Second})
	if err != nil {
		return rabbitConnectionChannelAndQueue, queueAndExchange, err
	}

	mqChannel, err = mqConnection.Channel()
	if err != nil {
		return rabbitConnectionChannelAndQueue, queueAndExchange, err
	}

	if connDetails.QueueAndExchange.ExchangeName == "" && connDetails.QueueAndExchange.QueueName == "" {
		msg = fmt.Sprintf("error : both (connDetails.Exchange) and (connDetails.QueueName) cannot be empty string, only one of them should be empty.")
		log.Printf(msg)
		return rabbitConnectionChannelAndQueue, queueAndExchange, err
	}

	if connDetails.QueueAndExchange.ExchangeName != "" && connDetails.QueueAndExchange.QueueName != "" {
		msg = fmt.Sprintf("error : both (connDetails.Exchange) and (connDetails.QueueName) cannot be non-empty string, only one of them should be empty.")
		log.Printf(msg)
		return rabbitConnectionChannelAndQueue, queueAndExchange, err
	}
	/* ************************************************************************************************************ */
	if connDetails.QueueAndExchange.ExchangeName != "" {
		err = mqChannel.ExchangeDeclare(
			connDetails.QueueAndExchange.ExchangeName, // name
			"fanout", // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		)
		if err != nil {
			return rabbitConnectionChannelAndQueue, queueAndExchange, err
		}

		rabbitConnectionChannelAndQueue = ConnectionChannelQueue{
			MQConnection: mqConnection,
			MQChannel:    mqChannel,
			MQQueue:      amqp.Queue{},
			Broadcast:    SendForAll,
		}
		queueAndExchange.ExchangeName = connDetails.QueueAndExchange.ExchangeName
	}
	/* ************************************************************************************************************ */
	if connDetails.QueueAndExchange.QueueName != "" {
		mqQueue, err = mqChannel.QueueDeclare(
			connDetails.QueueAndExchange.QueueName, // name
			true,                                   // durable
			false,                                  // delete when unused
			false,                                  // exclusive
			false,                                  // no-wait
			nil,                                    // arguments
		)
		if err != nil {
			return rabbitConnectionChannelAndQueue, queueAndExchange, err
		}

		rabbitConnectionChannelAndQueue = ConnectionChannelQueue{
			MQConnection: mqConnection,
			MQChannel:    mqChannel,
			MQQueue:      mqQueue,
			Broadcast:    SendForSession,
		}

		queueAndExchange.QueueName = connDetails.QueueAndExchange.QueueName
	}

	/* ************************************************************************************************************ */
	return rabbitConnectionChannelAndQueue, queueAndExchange, err
}

func publishMessageGeneric(rabbitConnectionChannelAndQueue ConnectionChannelQueue, queueAndExchange QueueAndExchange, data interface{}) {
	defer elapsed("__FUNC__: publishMessageGeneric")()
	body := ""
	msg := ""
	jsonStr, err := json.Marshal(data)
	if err != nil {
		log.Errorf("could not marshal json data : %v", err.Error())
	} else {
		body = string(jsonStr)
		log.Printf("body : (%v)", body)
	}

	if rabbitConnectionChannelAndQueue.Broadcast == SendForSession {
		err = rabbitConnectionChannelAndQueue.MQChannel.Publish(
			"",                         // exchange
			queueAndExchange.QueueName, // routing key
			false,                      // mandatory
			false,                      // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		if err != nil {
			msg = fmt.Sprintf("error : publishMessageGeneric() : SendForSession : could not publish the message : %v", err.Error())
			log.Printf(msg)
			return
		}
		log.Printf("# [x] Sent %s\n", body)
	}

	if rabbitConnectionChannelAndQueue.Broadcast == SendForAll {
		err = rabbitConnectionChannelAndQueue.MQChannel.Publish(
			queueAndExchange.ExchangeName, // exchange
			"",                            // routing key
			false,                         // mandatory
			false,                         // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		if err != nil {
			msg = fmt.Sprintf("error : publishMessageGeneric() : SendForAll : could not publish the message : %v", err.Error())
			log.Printf(msg)
			return
		}
	}

}

func pushMessagesGeneric(queueAndExchange QueueAndExchange, rabbitConnectionChannelAndQueue ConnectionChannelQueue) {
	defer elapsed("__FUNC__: pushMessagesGeneric")()
	done := make(chan bool)
	myMap := make(map[string]string)

	log.Printf("will start pushing messages onto rabbitmq ...")

	go func() {
		start := time.Now()
		counter := 0
		for {
			counter++

			time.Sleep(200 * time.Millisecond)

			randomStr := getRandomString()
			randomUUID := getRandomUUID()

			myMap["random_string"] = randomStr
			myMap["random_uuid"] = randomUUID

			if rabbitConnectionChannelAndQueue.Broadcast == SendForSession {
				myMap["data"] = fmt.Sprintf("QUEUE_%v_%v", queueAndExchange.QueueName, counter)
			}

			if rabbitConnectionChannelAndQueue.Broadcast == SendForAll {
				myMap["data"] = fmt.Sprintf("EXCHANGE_%v_%v", queueAndExchange.ExchangeName, counter)
			}

			publishMessageGeneric(rabbitConnectionChannelAndQueue, queueAndExchange, myMap)

			duration := time.Since(start)

			totalTimeElapsed := duration.Seconds()

			log.Printf("totalTimeElapsed : %v", totalTimeElapsed)

			if totalTimeElapsed > 6 {
				done <- true
			}
		}
	}()

	log.Printf("waiting for done <- true")

	<-done

	log.Printf("@ Done sending !")
}

/*
curl -v -i -u $MQUSER:$MQPASS \
-H "accept:application/json" \
-H "content-type:application/json" \
-X PUT -d '{"type":"fanout","durable":true}' \
http://$MQHOST:15672/api/exchanges/%2F/my-new-exchange
*/
