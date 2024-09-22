package models

type Row struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	PlaceID int    `json:"place_id"` // ID места
}
