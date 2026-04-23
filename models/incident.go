package models

import (
	"time"
)

//Incident holds the information about the incident
type Incident struct {
	ID              int                 `json:"id"`
	Time            time.Time           `json:"time"`
	Title           string              `json:"title"`
	Status          IncidentStatus      `json:"status"`
	Services        []ServiceUpdate     `json:"services,omitempty"`
	Updates         []*StatusUpdate     `json:"updates"`
	UpdatedAt       time.Time           `json:"updatedAt"`
	Maintenance     IncidentMaintenance `json:"maintenance"`   // Deprecated
	IsMaintenance   bool                `json:"isMaintenance"` // Deprecated
	OriginalTweetID int64               `json:"originalTweetId"`
	LatestTweetID   int64               `json:"latestTweetId"`
	Labels          map[string]string   `json:"labels,omitempty"`
}

//IncidentMaintenance contains the data about a scheduled maintenance.
type IncidentMaintenance struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

//IncidentStatus represents the status of the incident
type IncidentStatus string

func (is IncidentStatus) String() string {
	return string(is)
}

const (
	//IncidentStatusScheduledMaintenance - Schedule maintenance Incident
	IncidentStatusScheduledMaintenance IncidentStatus = "Scheduled Maintenance" // deprecated
	//IncidentStatusInvestigating - Investigating Incident
	IncidentStatusInvestigating IncidentStatus = "Investigating"
	//IncidentStatusIdentified - Identified cause of Incident
	IncidentStatusIdentified IncidentStatus = "Identified"
	//IncidentStatusUpdate - An update to the Incident. Does not update overall incident status
	IncidentStatusUpdate IncidentStatus = "Update"
	//IncidentStatusMonitoring - Monitoring typically after a fix applied
	IncidentStatusMonitoring IncidentStatus = "Monitoring"
	//IncidentStatusResolved - Resolved the incident
	IncidentStatusResolved IncidentStatus = "Resolved"
	//IncidentDefaultStatus is the default status of an incident
	IncidentDefaultStatus IncidentStatus = IncidentStatusInvestigating
)

//IncidentStatuses holds all of the valid incident statuses
var IncidentStatuses = map[string]IncidentStatus{
	"scheduled maintenance": IncidentStatusScheduledMaintenance, // deprecated
	"investigating":         IncidentStatusInvestigating,
	"identified":            IncidentStatusIdentified,
	"update":                IncidentStatusUpdate,
	"monitoring":            IncidentStatusMonitoring,
	"resolved":              IncidentStatusResolved,
}

var IncidentStatusArray = []IncidentStatus{
	IncidentStatusScheduledMaintenance, // deprecated
	IncidentStatusInvestigating,
	IncidentStatusIdentified,
	IncidentStatusUpdate,
	IncidentStatusMonitoring,
	IncidentStatusResolved,
}
