package models

type Order struct {
	ID       uint   `json:"id" gorm:"primary_key"`
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
	Notes    string `json:"notes"`
	Status   string `json:"status"`
}
