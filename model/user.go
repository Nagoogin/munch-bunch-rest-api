package model

type User struct {
	Username 	string 	`json:"username"`
	Fname		string	`json:"fname"`
	Lname		string	`json:"lname"`
	Email		string	`json:"email"`
	Cell		string	`json:"cell"`
}