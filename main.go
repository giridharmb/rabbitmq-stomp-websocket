package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"sync"
	"time"
)

func failOnError(err error, msg string) bool {
	if err != nil {
		log.Errorf("%v : %v", msg, err.Error())
		return true
	}
	return false
}

var (
	// environment variables : 'MQUSER' , 'MQPASS', 'MQHOST'
	// please set them !

	MQUSER string = "" // MQ Username
	MQPASS string = "" // MQ Password
	MQHOST string = "" // MQ Hostname , Ex: my-mq-server.company.com

	PGHOST string = "" // PostgreSQL Host
	PGUSER string = "" // PostgreSQL User
	PGPASS string = "" // PostgreSQL Pass
	PGDB   string = "" // PostgreSQL Database
)

func initApp() bool {

	// check if ENV Variables are set.

	// ---- rabitmq details ----

	MQUSER = os.Getenv("MQUSER")
	if MQUSER == "" {
		log.Error("please set 'MQUSER' environment variable !")
		return false
	}

	MQPASS = os.Getenv("MQPASS")
	if MQPASS == "" {
		log.Error("please set 'MQPASS' environment variable !")
		return false
	}

	MQHOST = os.Getenv("MQHOST")
	if MQHOST == "" {
		log.Error("please set 'MQHOST' environment variable !")
		return false
	}

	// ---- postgresql details ----

	PGHOST = os.Getenv("TEST_PGSQL_HOST")
	if PGHOST == "" {
		log.Error("please set 'TEST_PGSQL_HOST' environment variable !")
		return false
	}

	PGUSER = os.Getenv("TEST_USER")
	if PGUSER == "" {
		log.Error("please set 'TEST_USER' environment variable !")
		return false
	}

	PGPASS = os.Getenv("TEST_PASS")
	if PGPASS == "" {
		log.Error("please set 'TEST_PASS' environment variable !")
		return false
	}

	PGDB = os.Getenv("TEST_DB")
	if PGDB == "" {
		log.Error("please set 'TEST_DB' environment variable !")
		return false
	}

	return true
}

func main() {

	errorMessage := ""

	var validOperations = []string{"push_messages_on_exchange", "api"}

	var operation string

	flag.StringVar(&operation, "o", "none", "Specify operation. Valid List : ['push_messages_on_exchange' , 'api']")

	flag.Usage = func() {
		fmt.Printf("\nUsage of our Program: \n\n")
		fmt.Printf("./go-project -o <operation>\n\n")

		fmt.Printf("Valid List of <operations>:\n")
		for _, operation := range validOperations {
			fmt.Printf("  %v\n", operation)
		}

	}

	flag.Parse()

	successfullyInitialized := initApp()
	if !successfullyInitialized {
		return
	}

	/* ******************************************************************************************************* */
	exchangeName := "ex_change1"

	vhostName := "v_host1"

	err := CreateVHost(vhostName)
	if err != nil {
		log.Printf(err.Error())
		return
	}

	err = CreateFanoutExchange(exchangeName, vhostName)
	if err != nil {
		log.Printf(err.Error())
		return
	}

	err = SetVHostPermissions(MQUSER, vhostName)
	if err != nil {
		log.Printf(err.Error())
		return
	}

	exchange = exchangeName
	vhost = vhostName

	/* ******************************************************************************************************* */

	if operation == "api" {

		router := SetupHTTPRoutes()

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				log.Printf("IsAMQPConnectionClosed : %v", IsAMQPConnectionClosed())
				time.Sleep(5 * time.Second)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("http server : listening on port 10000 ...")
			log.Fatal(http.ListenAndServe(":10000", router))
		}()

		log.Printf("program will continue to serve http web server...")

		wg.Wait()

	} else if operation == "push_messages_on_exchange" {

		log.Printf(">> (MQ) : OPENING CONNECTION...")

		var queueAndExchange = QueueAndExchange{
			QueueName:    "",
			ExchangeName: exchange,
		}

		connDetails := ConnectionDetails{
			MQUser:           MQUSER,
			MQPass:           MQPASS,
			MQHost:           MQHOST,
			QueueAndExchange: queueAndExchange,
		}

		connChanQueue, queueAndExchange, err := connDetails.OpenConnection()
		if err != nil {
			log.Printf(">> (MQ) : error : %v", err.Error())
			errorMessage = fmt.Sprintf("HandlerBroadcast : could not create MQ connection for exchange  : %v", err.Error())
			log.Printf(errorMessage)
		}

		pushMessagesGeneric(queueAndExchange, connChanQueue)

	}
}
