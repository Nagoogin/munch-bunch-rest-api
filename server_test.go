package main

import (
	// "fmt"
	"log"
	"os"
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