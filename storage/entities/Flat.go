package entities

type Flat struct {
	Number int    `json:"number"`
	Price  int    `json:"price"`
	Rooms  int    `json:"rooms"`
	HomeId int    `json:"homeId"`
	Status string `json:"status"`
}
