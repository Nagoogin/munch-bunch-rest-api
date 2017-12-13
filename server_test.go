package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

var a App

const TABLE_CREATION_QUERY = `CREATE TABLE IF NOT EXISTS trucks
(
id SERIAL,
name TEXT NOT NULL,
CONSTRAINT trucks_pkey PRIMARY KEY (id)
)`

func checkTableExists() {
    if _, err := a.DB.Exec(TABLE_CREATION_QUERY); err != nil {
        log.Fatal(err)
    }
}

func clearTable() {
    a.DB.Exec("DELETE FROM trucks")
    a.DB.Exec("ALTER SEQUENCE trucks_id_seq RESTART WITH 1")
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

func addTrucks(count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO trucks(name) VALUES($1)", "Truck " + strconv.Itoa(i))
	}
}

func TestMain(m *testing.M) {
	a = App{}
	a.Initialize(
		os.Getenv("TEST_DB_USERNAME"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_NAME"))
	checkTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/api/v1/trucks", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentTruck(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/api/v1/truck/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Truck not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Truck not found'. Got '%s'", m["error"])
	}
}

func TestGetTruck(t *testing.T) {
	clearTable()
	addTrucks(1)

	req, _ := http.NewRequest("GET", "/api/v1/truck/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestGetTrucks(t *testing.T) {
	clearTable()
	addTrucks(10)

	req, _ := http.NewRequest("GET", "/api/v1/trucks", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestCreateTruck(t *testing.T) {
	clearTable()

	payload := []byte(`{"name":"test truck"}`)
	req, _ := http.NewRequest("POST", "/api/v1/truck", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test truck" {
		t.Errorf("Expected truck name to be 'test truck'. Got '%v'", m["name"])
	}

	if m["id"] != 1.0 {
		t.Errorf("Expected truck ID to be '1'. Got '%v'", m["id"])
	}
}

func TestUpdateTruck(t *testing.T) {
	clearTable()
	addTrucks(1)

	req, _ := http.NewRequest("GET", "/api/v1/truck/1", nil)
	response := executeRequest(req)
	var originalTruck map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalTruck)

	payload := []byte(`{"name":"Updated truck 1"}`)

	req, _ = http.NewRequest("PUT", "/api/v1/truck/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalTruck["id"] {
		t.Errorf("Expected the id to remain the unchanged (%v). Got %v", originalTruck["id"], m["id"])
	}

	if m["name"] == originalTruck["name"] {
		t.Errorf("Expected the name to change from '%v' to 'Updated truck 1'. Got '%v'", originalTruck["name"], m["name"])
	}
}

func TestDeleteTruck(t *testing.T) {
	clearTable()
	addTrucks(1)

	req, _ := http.NewRequest("GET", "/api/v1/truck/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/v1/truck/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/api/v1/truck/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code) 
}