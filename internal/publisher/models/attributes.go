package models

type Attributes struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	Altitude  float32 `json:"altitude,omitempty"`
	Accuracy  float32 `json:"gps_accuracy,omitempty"`
}
