package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() *mux.Router {
	router := mux.NewRouter()
	router.Use(accessControlMiddleware)
	router.HandleFunc("/", app.home)
	router.Path("/rest").Queries("time", "{time}").Queries("price", "{price}").Methods("GET").HandlerFunc(app.restaurants)
	router.Path("/rest").Queries("price", "{price}").Methods("GET").HandlerFunc(app.restaurants)
	router.Path("/rest").Queries("time", "{time}").Methods("GET").HandlerFunc(app.restaurants)
	router.Path("/rest").Queries("idrest", "{idrest}").Methods("GET").HandlerFunc(app.restaurants)
	router.Path("/rest").Methods("GET").HandlerFunc(app.restaurants)
	router.Path("/rest_info").Methods("GET").HandlerFunc(app.restaurantsInfo)
	router.Path("/free_table_rest").Methods("GET").HandlerFunc(app.freeRest)
	router.Path("/free_table_rest").Queries("idrest", "{idrest}").Methods("GET").HandlerFunc(app.freeRest)
	router.Path("/reserve").Methods("POST").HandlerFunc(app.reserveRest)     // резерв столика
	router.Path("/reserve").Methods("DELETE").HandlerFunc(app.deleteReserve) // удаление резерва

	http.Handle("/", router)

	return router
}

// access control and  CORS middleware
func accessControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS,PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}
