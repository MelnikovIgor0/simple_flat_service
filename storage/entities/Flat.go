package entities

type Flat struct {
	Number int    `json:"id"`
	Price  int    `json:"price"`
	Rooms  int    `json:"rooms"`
	HomeId int    `json:"house_id"`
	Status string `json:"status"`
}
