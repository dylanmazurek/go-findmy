package models

type SemanticLocation struct {
	Names     []string `json:"names"`
	Longitude float64  `json:"longitude"`
	Latitude  float64  `json:"latitude"`
}
