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
	"github.com/nagoogin/munch-bunch-rest-api/db"
	// "github.com/dimiro1/health"
	// "github.com/dimiro1/health/db"
	// "github.com/dimiro1/health/url"
	"github.com/nagoogin/munch-bunch-rest-api/handler"

	_ "github.com/lib/pq"
)

// PostgreSQL DB info
const (
	HOST		= "localhost"
	PORT		= 5432
	USER		= "kevinnguyen"
	DBNAME		= "munchbunch"
)

type App struct {
	Router 		*mux.Router
	Subrouter 	*mux.Router
	DB 			*sql.DB
}

func (a *App) Initialize(user, password, dbname string) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
    HOST, PORT, USER, DBNAME)

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
	a.Subrouter.Methods("GET").Path("/truck/{id:[0-9]+}").HandlerFunc(a.getTruck)
	a.Subrouter.Methods("GET").Path("/trucks").HandlerFunc(a.getTrucks)
	a.Subrouter.Methods("POST").Path("/truck").HandlerFunc(a.createTruck)
	a.Subrouter.Methods("PUT").Path("/truck/{id:[0-9]+}").HandlerFunc(a.updateTruck)
	a.Subrouter.Methods("DELETE").Path("/truck/{id:[0-9]+}").HandlerFunc(a.deleteTruck)
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

func (a *App) getTruck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid truck ID")
		return
	}

	t := db.Truck{ID: id}
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

	trucks, err := db.GetTrucks(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, trucks)
}

func (a *App) createTruck(w http.ResponseWriter, r *http.Request) {
	var t db.Truck
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
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var t db.Truck
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
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	t := db.Truck{ID: id}
	if err := t.DeleteTruck(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func main() {
	a := App{}
    a.Initialize(
        os.Getenv("APP_DB_USERNAME"),
        os.Getenv("APP_DB_PASSWORD"),
        os.Getenv("APP_DB_NAME"))

    a.Run(":8080")

	// // Test connection to PSQL DB
	// err = database.Ping()  
	// if err != nil {  
	//   log.Fatal(err)
	// }
	// fmt.Println("Successfully opened connection")

	// psqlChecker := db.NewPostgreSQLChecker(database)

	// r := mux.NewRouter()
	// r.Path("/api/v1").HandlerFunc(handler.StatusHandler)

	// s := r.PathPrefix("/api/v1").Subrouter()

	// healthHandler := health.NewHandler()
	// healthHandler.AddChecker("api", url.NewChecker("http://localhost:8080/api/v1"))
	// healthHandler.AddChecker("db", psqlChecker)

	// s.Methods("GET").Path("/trucks").HandlerFunc(handler.GetTrucks)
	// s.Methods("GET").Path("/trucks/{name}").HandlerFunc(handler.GetTruck)
	// s.Methods("DELETE").Path("/trucks/{name}").HandlerFunc(handler.DeleteTruck)

	// s.Path("/health").Handler(healthHandler)
}