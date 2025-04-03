package models

type Report struct {
	UniqueId  string  `json:"unique_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude,omitempty"`
	Accuracy  float64 `json:"gps_accuracy,omitempty"`
}
