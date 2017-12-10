package model

type User struct {
	username 	string 	`json:"username"`
	fname		string	`json:"fname"`
	lname		string	`json:"lname"`
	email		string	`json:"email"`
	cell		string	`json:"cell"`
}