package models

type Device struct {
	UUID string
	Name string

	LocationReports []LocationReport
}
