package main

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
)

var controller = &Controller{}

const (
	updateevent = "/updateevent"
	getevents = "/getevents"
	addevent ="/addevent"
)

// Route defines a route
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes defines the list of routes of our API
type Routes []Route

// Initialize the routes.
var NonTlsroutes = Routes {
	Route {
		"AddEvent",
		"POST",
		addevent,
		controller.AddEvent,
	},
	Route {
		"UpdateEvent",
		"PUT",
		updateevent,
		controller.UpdateEvent,
	},
	Route {
		"GetEvents",
		"GET",
		getevents,
		controller.GetEvents,
	},
}

// NewRouter configures a new router to the API
func NewNonTLSRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range NonTlsroutes	 {
		var handler http.Handler
		log.Println(route.Name)
		handler = route.HandlerFunc

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	return router
}
