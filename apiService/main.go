package main

import (
	"crypto/tls"
	"github.com/gorilla/handlers"
	"log"
	"net/http"
	"os"
)

func main(){
	log.Println("starting api microservice")

	log.Println("Exposing the following REST APIs")

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort="80"
	}

	log.Print("Setting http port:",httpPort)

	//Defining accesses
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET"})

	// create routes for non tls
	routertlsfree := NewNonTLSRouter()

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	// Create a Server instance to listen on port 443 with the TLS config
	server := &http.Server{
		Addr:      ":"+httpPort,
		TLSConfig: conf,
		Handler: handlers.CORS(allowedOrigins, allowedMethods)(routertlsfree),
	}

	//Start listening on the origin and methods
	log.Fatal(server.ListenAndServe())
}
