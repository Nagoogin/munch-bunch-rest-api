package main

import (
	"bytes"
	"encoding/json"
	"os"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"github.com/Nagoogin/munch-bunch-rest-api/crypto"
)

var a App

func clearTableTrucks() {
    a.DB.Exec("DELETE FROM trucks")
    a.DB.Exec("ALTER SEQUENCE trucks_id_seq RESTART WITH 1")
}

func clearTableUsers() {
	a.DB.Exec("DELETE FROM users")
	a.DB.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func addUsers(count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO users(username, hash, fname, lname, email, hasTruck) VALUES($1, $2, $3, $4, $5, $6)",
			"User" + strconv.Itoa(i), crypto.HashAndSalt([]byte("password")), "first-name", "last-name", "email@test.com", false)
	}
}

func addTrucks(count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO trucks(name) VALUES($1)", "Truck " + strconv.Itoa(i))
	}
}

func getJWT() string {
	clearTableUsers()
	addUsers(1)

	payload := []byte(`{"username":"User0","password":"password"}`)
	req, _ := http.NewRequest("POST", "/api/v1/auth/authenticate", bytes.NewBuffer(payload))
	response := executeRequest(req)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	return "Bearer " + m["data"].(map[string]interface{})["token"].(string)
}

func TestMain(m *testing.M) {
	a = App{}
	a.Initialize(
		os.Getenv("TEST_DB_USERNAME"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_NAME"))
	a.CheckTablesExist()
	code := m.Run()
	clearTableTrucks()
	os.Exit(code)
}

// Auth endpoint tests

func TestAuthenticate(t *testing.T) {
	clearTableUsers()
	addUsers(1)

	payload := []byte(`{"username":"User0","password":"password"}`)
	req, _ := http.NewRequest("POST", "/api/v1/auth/authenticate", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["token"] == "" {
		t.Errorf("Expected a JWT as the response, got '%s'", m["error"])
	}
}

func TestValidationMiddleware(t *testing.T) {
	clearTableTrucks()
	addTrucks(1)

	jwt := getJWT()
	req, _ := http.NewRequest("GET", "/api/v1/truck/1", nil)
	req.Header.Set("Authorization", jwt)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}

// User endpoint tests

func TestGetNonExistentUser(t *testing.T) {
	clearTableUsers()

	req, _ := http.NewRequest("GET", "/api/v1/user/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["message"] != "User not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'User not found'. Got '%s'", m["message"].(string))
	}
}

func TestGetUser(t *testing.T) {
	clearTableUsers()
	addUsers(1)

	req, _ := http.NewRequest("GET", "/api/v1/user/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestCreateUser(t *testing.T) {
	clearTableUsers()

	payload := []byte(`{"username":"User1","hash":"password","fname":"first-name","lname":"last-name","email":"email@test.com"}`)
	req, _ := http.NewRequest("POST", "/api/v1/user", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["data"].(map[string]interface{})["username"] != "User1" {
		t.Errorf("Expected username to be 'User1'. Got '%v'", m["data"].(map[string]interface{})["username"])
	}
	if !crypto.ComparePasswords(m["data"].(map[string]interface{})["hash"].(string), []byte("password")) {
		t.Errorf("Unexpected password")
	}
	if m["data"].(map[string]interface{})["fname"] != "first-name" {
		t.Errorf("Expected fname to be 'first-name'. Got '%v'", m["data"].(map[string]interface{})["fname"])
	}
	if m["data"].(map[string]interface{})["lname"] != "last-name" {
		t.Errorf("Expected lname to be 'last-name'. Got '%v'", m["data"].(map[string]interface{})["lname"])
	}
	if m["data"].(map[string]interface{})["email"] != "email@test.com" {
		t.Errorf("Expected email to be 'email@test.com'. Got '%v'", m["data"].(map[string]interface{})["email"])
	}
	if m["data"].(map[string]interface{})["id"] != 1.0 {
		t.Errorf("Expected truck ID to be '1'. Got '%v'", m["data"].(map[string]interface{})["id"])
	}
}

func TestUpdateUser(t *testing.T) {
	clearTableUsers()
	addUsers(1)

	req, _ := http.NewRequest("GET", "/api/v1/user/1", nil)
	response := executeRequest(req)
	var originalUser map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalUser)

	payload := []byte(`{"username":"Updated1","hash":"updated-password","fname":"updated-first-name","lname":"updated-last-name","email":"updated.email@test.com"}`)

	req, _ = http.NewRequest("PUT", "/api/v1/user/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["data"].(map[string]interface{})["id"] != originalUser["data"].(map[string]interface{})["id"] {
		t.Errorf("Expected the id to remain the unchanged (%v). Got %v", originalUser["data"].(map[string]interface{})["id"], m["data"].(map[string]interface{})["id"])
	}
	if m["data"].(map[string]interface{})["hash"] == originalUser["data"].(map[string]interface{})["hash"] {
		t.Errorf("Expected hash to change from '%v' to '%v'. Got '%v'", originalUser["data"].(map[string]interface{})["hash"], crypto.HashAndSalt([]byte("updated-password")), m["data"].(map[string]interface{})["hash"])
	}
	if m["data"].(map[string]interface{})["username"] == originalUser["data"].(map[string]interface{})["username"] {
		t.Errorf("Expected the username to change from '%v' to 'Updated1'. Got '%v'", originalUser["data"].(map[string]interface{})["username"], m["data"].(map[string]interface{})["username"])
	}
	if m["data"].(map[string]interface{})["fname"] == originalUser["data"].(map[string]interface{})["fname"] {
		t.Errorf("Expected the fname to change from '%v' to 'updated-first-name'. Got '%v'", originalUser["data"].(map[string]interface{})["fname"], m["data"].(map[string]interface{})["fname"])
	}
	if m["data"].(map[string]interface{})["lname"] == originalUser["data"].(map[string]interface{})["lname"] {
		t.Errorf("Expected the lname to change from '%v' to 'updated-last-name'. Got '%v'", originalUser["data"].(map[string]interface{})["lname"], m["data"].(map[string]interface{})["lname"])
	}
	if m["data"].(map[string]interface{})["email"] == originalUser["data"].(map[string]interface{})["email"] {
		t.Errorf("Expected the email to change from '%v' to 'updated.email@test.com'. Got '%v'", originalUser["data"].(map[string]interface{})["email"], m["data"].(map[string]interface{})["email"])
	}
}

func TestDeleteUser(t *testing.T) {
	clearTableUsers()
	addUsers(1)

	req, _ := http.NewRequest("GET", "/api/v1/user/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/v1/user/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/api/v1/user/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code) 
}

// Truck endpoint tests

func TestEmptyTable(t *testing.T) {
	clearTableTrucks()

	jwt := getJWT()
	req, _ := http.NewRequest("GET", "/api/v1/trucks", nil)
	req.Header.Set("Authorization", jwt)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if len(m["data"].([]interface{})) > 0 {
		t.Errorf("Expected an empty array. Got %s", m["data"])
	}
}

func TestGetNonExistentTruck(t *testing.T) {
	clearTableTrucks()

	jwt := getJWT()
	req, _ := http.NewRequest("GET", "/api/v1/truck/1", nil)
	req.Header.Set("Authorization", jwt)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["message"].(string) != "Truck not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Truck not found'. Got '%s'", m["message"].(string))
	}
}

func TestGetTruck(t *testing.T) {
	clearTableTrucks()
	addTrucks(1)

	jwt := getJWT()
	req, _ := http.NewRequest("GET", "/api/v1/truck/1", nil)
	req.Header.Set("Authorization", jwt)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestGetTrucks(t *testing.T) {
	clearTableTrucks()
	addTrucks(10)

	jwt := getJWT()
	req, _ := http.NewRequest("GET", "/api/v1/trucks", nil)
	req.Header.Set("Authorization", jwt)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestCreateTruck(t *testing.T) {
	clearTableTrucks()

	jwt := getJWT()

	payload := []byte(`{"name":"test truck"}`)
	req, _ := http.NewRequest("POST", "/api/v1/truck", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", jwt)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["data"].(map[string]interface{})["name"] != "test truck" {
		t.Errorf("Expected truck name to be 'test truck'. Got '%v'", m["data"].(map[string]interface{})["name"])
	}
	if m["data"].(map[string]interface{})["id"] != 1.0 {
		t.Errorf("Expected truck ID to be '1'. Got '%v'", m["data"].(map[string]interface{})["id"])
	}
}

func TestUpdateTruck(t *testing.T) {
	clearTableTrucks()
	addTrucks(1)

	jwt := getJWT()
	req, _ := http.NewRequest("GET", "/api/v1/truck/1", nil)
	req.Header.Set("Authorization", jwt)
	response := executeRequest(req)
	var originalTruck map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalTruck)

	payload := []byte(`{"name":"Updated truck 1"}`)

	req, _ = http.NewRequest("PUT", "/api/v1/truck/1", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", jwt)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["data"].(map[string]interface{})["id"] != originalTruck["data"].(map[string]interface{})["id"] {
		t.Errorf("Expected the id to remain the unchanged (%v). Got %v", originalTruck["data"].(map[string]interface{})["id"], m["data"].(map[string]interface{})["id"])
	}
	if m["data"].(map[string]interface{})["name"] == originalTruck["data"].(map[string]interface{})["name"] {
		t.Errorf("Expected the name to change from '%v' to 'Updated truck 1'. Got '%v'", originalTruck["data"].(map[string]interface{})["name"], m["data"].(map[string]interface{})["name"])
	}
}

func TestDeleteTruck(t *testing.T) {
	clearTableTrucks()
	addTrucks(1)

	jwt := getJWT()
	req, _ := http.NewRequest("GET", "/api/v1/truck/1", nil)
	req.Header.Set("Authorization", jwt)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/v1/truck/1", nil)
	req.Header.Set("Authorization", jwt)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/api/v1/truck/1", nil)
	req.Header.Set("Authorization", jwt)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code) 
}
