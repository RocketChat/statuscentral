package models

import (
	"time"
)

// AggregatedIncident contains the incidents aggregated
type AggregatedIncident struct {
	Time      time.Time
	Incidents []*Incident
}

// AggregatedIncidents holds several aggregated incidents
type AggregatedIncidents []AggregatedIncident

// ScheduledMaintenanceAggregatedByDay contains the scheduled maintenance aggregated by day
type ScheduledMaintenanceAggregatedByDay struct {
	Time                 time.Time
	ScheduledMaintenance []*ScheduledMaintenance
}

// AggregatedScheduledMaintenances holds several aggregated scheduled maintenances
type AggregatedScheduledMaintenances struct {
	Days  []ScheduledMaintenanceAggregatedByDay
	Count int
}
