package model

type Truck struct {
	Name 	string 	`json:"name"`
	Cell	string	`json:"cell"`
	Address string 	`json:"address"`
	City	string 	`json:"city"`
	State	string	`json:"state"`
	Zip		string	`json:"zip"`
}