package main 

import (
	"fmt"
	"log"
	"net/http"
	"database/sql"
	"os"

	"github.com/gorilla/mux"
	// "github.com/dimiro1/health"
	// "github.com/dimiro1/health/db"
	// "github.com/dimiro1/health/url"
	// "github.com/nagoogin/munch-bunch-rest-api/handler"

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
	Router 	*mux.Router
	DB 		*sql.DB
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
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(":8080", a.Router))
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