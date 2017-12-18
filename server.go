package main 

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"database/sql"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/dimiro1/health"
	"github.com/dimiro1/health/db"
	"github.com/dimiro1/health/url"
	"github.com/Nagoogin/munch-bunch-rest-api/database"
	"github.com/Nagoogin/munch-bunch-rest-api/handler"

	_ "github.com/lib/pq"
)

type App struct {
	Router 		*mux.Router
	Subrouter 	*mux.Router
	DB 			*sql.DB
}

// TODO place contants in a constants.go file
const USER_TABLE_CREATION_QUERY = `CREATE TABLE IF NOT EXISTS users
(
id SERIAL,
username TEXT NOT NULL,
fname TEXT NOT NULL,
lname TEXT NOT NULL,
email TEXT NOT NULL,
CONSTRAINT users_pkey PRIMARY KEY (id)
)`

const TRUCK_TABLE_CREATION_QUERY = `CREATE TABLE IF NOT EXISTS trucks
(
id SERIAL,
name TEXT NOT NULL,
CONSTRAINT trucks_pkey PRIMARY KEY (id)
)`

func (a *App) CheckTablesExist() {
    if _, err := a.DB.Exec(USER_TABLE_CREATION_QUERY); err != nil {
        log.Fatal(err)
    }
    if _, err2 := a.DB.Exec(TRUCK_TABLE_CREATION_QUERY); err2 != nil {
    	log.Fatal(err2)
    }
}

func (a *App) Initialize(user, password, dbname string) {
	psqlInfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
    user, password, dbname)

	var err error
    a.DB, err = sql.Open("postgres", psqlInfo)  
	if err != nil {  
	  log.Fatal(err)
	}

	a.Router = mux.NewRouter();
	a.Subrouter = a.Router.PathPrefix("/api/v1").Subrouter()
	a.InitializeRoutes()
}

func (a *App) InitializeRoutes() {
	a.Router.Path("/api/v1").HandlerFunc(handler.StatusHandler)

	// User endpoints
	a.Subrouter.Methods("GET").Path("/user/{id:[0-9]+}").HandlerFunc(a.getUser)
	a.Subrouter.Methods("POST").Path("/user").HandlerFunc(a.createUser)
	a.Subrouter.Methods("PUT").Path("/user/{id:[0-9]+}").HandlerFunc(a.updateUser)
	a.Subrouter.Methods("DELETE").Path("/user/{id:[0-9]+}").HandlerFunc(a.deleteUser)

	// Truck endpoints
	a.Subrouter.Methods("GET").Path("/truck/{id:[0-9]+}").HandlerFunc(a.getTruck)
	a.Subrouter.Methods("GET").Path("/trucks").HandlerFunc(a.getTrucks)
	a.Subrouter.Methods("POST").Path("/truck").HandlerFunc(a.createTruck)
	a.Subrouter.Methods("PUT").Path("/truck/{id:[0-9]+}").HandlerFunc(a.updateTruck)
	a.Subrouter.Methods("DELETE").Path("/truck/{id:[0-9]+}").HandlerFunc(a.deleteTruck)

	psqlChecker := db.NewPostgreSQLChecker(a.DB)
	healthHandler := health.NewHandler()
	healthHandler.AddChecker("api", url.NewChecker("http://localhost:8080/api/v1"))
	healthHandler.AddChecker("db", psqlChecker)
	a.Subrouter.Path("/health").Handler(healthHandler)
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	u := database.User{ID: id}
	if err := u.GetUser(a.DB); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var u database.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := u.CreateUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, u)
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var u database.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	
	u.ID = id
	if err := u.UpdateUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	u := database.User{ID: id}
	if err := u.DeleteUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) getTruck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid truck ID")
		return
	}

	t := database.Truck{ID: id}
	if err := t.GetTruck(a.DB); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Truck not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, t)
}

func (a *App) getTrucks(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}

	if start < 0 {
		start = 0
	}

	trucks, err := database.GetTrucks(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, trucks)
}

func (a *App) createTruck(w http.ResponseWriter, r *http.Request) {
	var t database.Truck
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := t.CreateTruck(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, t)
}

func (a *App) updateTruck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid truck ID")
		return
	}

	var t database.Truck
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	
	t.ID = id
	if err := t.UpdateTruck(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, t)
}

func (a *App) deleteTruck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid truck ID")
		return
	}

	t := database.Truck{ID: id}
	if err := t.DeleteTruck(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func main() {
	// certPath := "server.pem"
	// keyPath := "server.key"

	a := App{}
    a.Initialize(
        os.Getenv("APP_DB_USERNAME"),
        os.Getenv("APP_DB_PASSWORD"),
        os.Getenv("APP_DB_NAME"))
    a.CheckTablesExist()
    a.Run(":8080")
}