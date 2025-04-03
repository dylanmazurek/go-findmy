package models

import (
	"fmt"
	"time"
)

type LocationReport struct {
	UniqueId *string

	ReportType ReportType
	ReportTime time.Time

	Latitude  float64
	Longitude float64
	Altitude  float64
	Accuracy  float64

	// Semantic location name
	SemanticName *string
}

type ReportType int8

const (
	ReportTypeSemantic ReportType = iota
	ReportTypeLocation
)

func (r *ReportType) String() string {
	switch *r {
	case ReportTypeSemantic:
		return "semantic"
	case ReportTypeLocation:
		return "location"
	default:
		return "unknown"
	}
}

func (l *LocationReport) String() string {
	reportType := l.ReportType

	var outputStr string
	switch reportType {
	case ReportTypeSemantic:
		outputStr = fmt.Sprintf("[%s] near: %s", l.ReportTime.Format(time.Stamp), *l.SemanticName)
	case ReportTypeLocation:
		outputStr = fmt.Sprintf("[%s] lat: %.6f lng: %.6f alt: %f", l.ReportTime.Format(time.Stamp), l.Latitude, l.Longitude, l.Altitude)
	default:
		return "unknown report type"
	}

	return outputStr
}
