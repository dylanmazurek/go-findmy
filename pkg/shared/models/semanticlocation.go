package models

type SemanticLocation struct {
	Names     []string `json:"names"`
	Longitude float32  `json:"longitude"`
	Latitude  float32  `json:"latitude"`
}
