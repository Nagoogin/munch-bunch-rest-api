package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

// User handlers

func GetUser()

// Food truck handlers

func GetTrucks(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func GetTruck(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func DeleteTruck(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func writeJsonResponse(w http.ResponseWriter, bytes []byte) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteBytes(bytes)
}
