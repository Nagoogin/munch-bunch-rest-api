package handler

import (
	//"encoding/json"
	//"github.com/gorilla/mux"
	//"io/ioutil"
	"net/http"
)

// Status handlers

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("API is up and running"))
}


