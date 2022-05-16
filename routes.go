package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/rs/cors"
	"log"
	"net/http"
)

var cookie *sessions.CookieStore

func InitializeSessionStore() {
	randStr := getRandomString()
	log.Printf("InitializeSessionStore() : randStr : %v", randStr)

	cookie = sessions.NewCookieStore([]byte(randStr))

}

func SetupHTTPRoutes() http.Handler {

	InitializeSessionStore()

	router := mux.NewRouter().StrictSlash(false)

	corsObject := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},                                                // All origins
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}, // Allowing only get, just an example
	})

	corsRouterHandler := corsObject.Handler(router)

	router.HandleFunc("/", Home).Methods("GET")
	router.HandleFunc("/broadcast", HandlerBroadcast).Methods("POST")
	router.HandleFunc("/api/v1/rabbitConnectionIsClosed", HandlerGetRabbitMQConnectionStatus).Methods("GET")

	return corsRouterHandler
}
