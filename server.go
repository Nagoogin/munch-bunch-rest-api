package main 

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"database/sql"
	"os"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/dimiro1/health"
	"github.com/dimiro1/health/db"
	"github.com/dimiro1/health/url"
	"github.com/Nagoogin/munch-bunch-rest-api/database"
	"github.com/Nagoogin/munch-bunch-rest-api/handler"
	"github.com/Nagoogin/munch-bunch-rest-api/crypto"

	_ "github.com/lib/pq"
)

type App struct {
	Router 		*mux.Router
	Subrouter 	*mux.Router
	DB 			*sql.DB
}

type JwtToken struct {
	Token string `json:"token"`
}

// TODO place contants in a constants.go file
const USER_TABLE_CREATION_QUERY = `CREATE TABLE IF NOT EXISTS users
(
id SERIAL,
username TEXT NOT NULL,
hash TEXT NOT NULL,
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

	// Auth endpoints
	a.Subrouter.Methods("POST").Path("/auth/register").HandlerFunc(a.Register)
	a.Subrouter.Methods("POST").Path("/auth/logout").HandlerFunc(a.Logout)
	a.Subrouter.Methods("POST").Path("/auth/authenticate").HandlerFunc(a.CreateToken)

	// User endpoints
	a.Subrouter.Methods("GET").Path("/user/{id:[0-9]+}").HandlerFunc(a.GetUser)
	a.Subrouter.Methods("POST").Path("/user").HandlerFunc(a.CreateUser)
	a.Subrouter.Methods("PUT").Path("/user/{id:[0-9]+}").HandlerFunc(a.UpdateUser)
	a.Subrouter.Methods("DELETE").Path("/user/{id:[0-9]+}").HandlerFunc(a.DeleteUser)

	a.Subrouter.Methods("GET").Path("/user{id:[0-9]+}/orders").HandlerFunc(a.GetOrdersForUser)

	// Truck endpoints
	a.Subrouter.Methods("GET").Path("/truck/{id:[0-9]+}").HandlerFunc(a.GetTruck)
	a.Subrouter.Methods("GET").Path("/trucks").HandlerFunc(a.GetTrucks)
	a.Subrouter.Methods("POST").Path("/truck").HandlerFunc(a.CreateTruck)
	a.Subrouter.Methods("PUT").Path("/truck/{id:[0-9]+}").HandlerFunc(a.UpdateTruck)
	a.Subrouter.Methods("DELETE").Path("/truck/{id:[0-9]+}").HandlerFunc(a.DeleteTruck)

	// Truck endpoints
	// a.Subrouter.Methods("GET").Path("/truck/{id:[0-9]+}").HandlerFunc(ValidateMiddleware(a.getTruck))
	// a.Subrouter.Methods("GET").Path("/trucks").HandlerFunc(ValidateMiddleware(a.getTrucks))
	// a.Subrouter.Methods("POST").Path("/truck").HandlerFunc(ValidateMiddleware(a.createTruck))
	// a.Subrouter.Methods("PUT").Path("/truck/{id:[0-9]+}").HandlerFunc(ValidateMiddleware(a.updateTruck))
	// a.Subrouter.Methods("DELETE").Path("/truck/{id:[0-9]+}").HandlerFunc(ValidateMiddleware(a.deleteTruck))

	a.Subrouter.Methods("GET").Path("/truck/{id:[0-9]+}/orders").HandlerFunc(a.GetOrdersForTruck)
	a.Subrouter.Methods("POST").Path("/truck/{id:[0-9]+}/orders").HandlerFunc(a.CreateOrderForTruck)
	a.Subrouter.Methods("PUT").Path("/truck/{id:[0-9]+}/order{id:[0-9]+}").HandlerFunc(a.UpdateOrderForTruck)
	a.Subrouter.Methods("DELETE").Path("/truck/{id:[0-9]+}/order/{id:[0-9]+}").HandlerFunc(a.DeleteOrderForTruck)


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

func (a *App) Register(w http.ResponseWriter, r *http.Request) {
	// TODO

	var u database.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// Creates and returns a JWT token if user credentials match those stored in the database
func (a *App) CreateToken(w http.ResponseWriter, r *http.Request) {

	// Read in user credentials from request body
	var userCred database.UserCredentials
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userCred); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Query user from database based on provided user credentials
	u := database.User{Username: userCred.Username}
	if err := u.GetUser(a.DB); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Compare stored hash with provided password from user credentials
	if crypto.ComparePasswords(u.Hash, []byte(userCred.Password)) {

		// If compare successful, create new JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims {
			"username": userCred.Username,
		})

		tokenString, err := token.SignedString([]byte("Secret"))
		if err != nil {
			log.Println(err)
		}
		respondWithJSON(w, http.StatusOK, JwtToken{Token: tokenString})
	}

	respondWithError(w, http.StatusForbidden, "Invalid password")
}

func ValidateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("authorization")
		if authorizationHeader != "" {
			bearerToken := strings.Split(authorizationHeader, " ")
			if len(bearerToken) == 2 {
                token, error := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
                    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                        return nil, fmt.Errorf("There was an error")
                    }
                    return []byte("secret"), nil
                })
                if error != nil {
                    respondWithError(w, http.StatusBadRequest, error.Error())
                    return
                }
                if token.Valid {
                    context.Set(r, "decoded", token.Claims)
                    next(w, r)
                } else {
                    respondWithError(w, http.StatusBadRequest, "Invalid authorization token")
                }
            }
        } else {
            respondWithError(w, http.StatusBadRequest, "An authorization header is required")
        }
	})
}

func (a *App) GetUser(w http.ResponseWriter, r *http.Request) {
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

func (a *App) CreateUser(w http.ResponseWriter, r *http.Request) {
	var u database.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Hash user password
	u.Hash = crypto.HashAndSalt([]byte(u.Hash))

	if err := u.CreateUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, u)
}

func (a *App) UpdateUser(w http.ResponseWriter, r *http.Request) {
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

func (a *App) DeleteUser(w http.ResponseWriter, r *http.Request) {
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

func (a *App) GetTruck(w http.ResponseWriter, r *http.Request) {
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

func (a *App) GetTrucks(w http.ResponseWriter, r *http.Request) {
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

func (a *App) CreateTruck(w http.ResponseWriter, r *http.Request) {
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

func (a *App) UpdateTruck(w http.ResponseWriter, r *http.Request) {
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

func (a *App) DeleteTruck(w http.ResponseWriter, r *http.Request) {
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

func (a *App) GetOrdersForUser(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (a *App) GetOrdersForTruck(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (a *App) CreateOrderForTruck(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (a *App) UpdateOrderForTruck(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (a *App) DeleteOrderForTruck(w http.ResponseWriter, r *http.Request) {
	// TODO
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