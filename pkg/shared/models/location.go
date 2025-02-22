package models

import (
	"fmt"
	"time"
)

type LocationReport struct {
	ReportType string
	ReportTime time.Time

	// Semantic location
	SemanticLocation string

	// Encrypted location
	Latitude  float64
	Longitude float64
	Altitude  float64
}

func (l *LocationReport) String() string {
	reportType := l.ReportType

	var outputStr string
	switch reportType {
	case "semantic":
		outputStr = fmt.Sprintf("[%s] near: %s", l.ReportTime.Format(time.Stamp), l.SemanticLocation)
	case "location":
		outputStr = fmt.Sprintf("[%s] lat: %.6f lng: %.6f alt: %f", l.ReportTime.Format(time.Stamp), l.Latitude, l.Longitude, l.Altitude)
	default:
		return "unknown report type"
	}

	return outputStr
}
