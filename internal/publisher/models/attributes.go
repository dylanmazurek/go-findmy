package models

type Attributes struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
	Altitude  float32 `json:"altitude"`
	Accuracy  float32 `json:"gps_accuracy"`
}
