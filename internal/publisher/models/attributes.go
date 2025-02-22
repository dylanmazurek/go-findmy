package models

type Attributes struct {
	Latitude    float32 `json:"latitude"`
	Longitude   float32 `json:"longitude"`
	GpsAccuracy float32 `json:"gps_accuracy"`
}
