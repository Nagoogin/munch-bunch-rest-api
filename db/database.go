package db

import (
	"database/sql"
)

type Truck struct {
	ID		int		`json:"id"`
	Name 	string 	`json:"name"`
	// Cell	string	`json:"cell"`
	// Address string 	`json:"address"`
	// City	string 	`json:"city"`
	// State	string	`json:"state"`
	// Zip		string	`json:"zip"`
}

func (t *Truck) GetTruck(db *sql.DB) error {
	return db.QueryRow("SELECT name FROM trucks WHERE id=$1", 
		t.ID).Scan(&t.Name)
}

func GetTrucks(db *sql.DB, start, count int) ([]Truck, error) { 
	rows, err := db.Query("SELECT id, name FROM trucks LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	trucks := []Truck{}

	for rows.Next() {
		var t Truck
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		trucks = append(trucks, t)
	}

	return trucks, nil
}

func (t *Truck) CreateTruck(db *sql.DB) error {
	err := db.QueryRow("INSERT INTO trucks(name) VALUES($1) RETURNING id",
		t.Name).Scan(&t.ID)

	if err != nil {
		return err
	}

	return nil
}

func (t *Truck) UpdateTruck(db *sql.DB) error {
	_, err := db.Exec("UPDATE trucks SET name=$1 WHERE id=$2",
		t.Name, t.ID)

	return err
}

func (t *Truck) DeleteTruck(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM trucks WHERE id=$1", t.ID)

	return err
}

