package main 

import (
	"github.com/gorilla/mux"
	"github.com/nagoogin/munch-bunch-rest-api/handler"
	"log"
	"net/http"
)

func main() {
	r := mux.NewRouter()
	s := r.PathPrefix("api/v1").Subrouter()

	sub.Methods("GET").Path("/trucks").HandlerFunc(handler.GetTrucks)
	sub.Methods("GET").Path("/trucks/{name}").HandlerFunc(handler.GetTruck)
	sub.Methods("DELETE").Path("/trucks/{name}").HandlerFunc(handler.DeleteTruck)

	log.Fatal(http.ListenAndServe(":8080", router))
}