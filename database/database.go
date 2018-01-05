package database

import (
	"database/sql"
)

type UserCredentials struct {
	Username 	string 	`json:"username"`
	Password	string	`json:"password"`
}

type User struct {
	ID 			int 	`json:"id"`
	Username 	string 	`json:"username"`
	Hash		string	`json:"hash"`
	Fname		string	`json:"fname"`
	Lname		string	`json:"lname"`
	Email		string	`json:"email"`
	HasTruck	bool	`json:"hasTruck"`
}

type Truck struct {
	ID		int		`json:"id"`
	Name 	string 	`json:"name"`
	// Cell	string	`json:"cell"`
	// Address string 	`json:"address"`
	// City	string 	`json:"city"`
	// State	string	`json:"state"`
	// Zip		string	`json:"zip"`
	// Locate	bool	`json:"locate"`
	// Location struct	`json:"location"`
}

type JsonRsp struct {
	Code 	int 		`json:"code"`
	Status	string		`json:"status"`
	Message	string		`json:"message"`
	Data 	interface{}	`json:"data"`
}

func (u *User) GetUser(db *sql.DB) error {
	return db.QueryRow("SELECT username, hash, fname, lname, email, hasTruck FROM users WHERE id=$1",
		u.ID).Scan(&u.Username, &u.Hash, &u.Fname, &u.Lname, &u.Email, &u.HasTruck)
}

func (u *User) GetUserByUsername(db *sql.DB) error {
	return db.QueryRow("SELECT username, hash, fname, lname, email, hasTruck FROM users WHERE username=$1",
		u.Username).Scan(&u.Username, &u.Hash, &u.Fname, &u.Lname, &u.Email, &u.HasTruck)
}

func (u *User) CreateUser(db *sql.DB) error {
	err := db.QueryRow("INSERT INTO users (username, hash, fname, lname, email, hasTruck) VALUES($1, $2, $3, $4, $5, $6) RETURNING id",
		u.Username, u.Hash, u.Fname, u.Lname, u.Email, u.HasTruck).Scan(&u.ID)

	if err != nil {
		return err
	}

	return nil
}

func (u *User) UpdateUser(db *sql.DB) error {
	_, err := db.Exec("UPDATE users SET username=$1, hash=$2, fname=$3, lname=$4, email=$5, hasTruck=$6 WHERE id=$7",
		u.Username, u.Hash, u.Fname, u.Lname, u.Email, u.HasTruck, u.ID)

	return err
}

func (u *User) DeleteUser(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM users WHERE id=$1", u.ID)

	return err
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
	err := db.QueryRow("INSERT INTO trucks (name) VALUES($1) RETURNING id",
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
