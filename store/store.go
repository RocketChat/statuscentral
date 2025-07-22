package store

import (
	"io"

	"github.com/RocketChat/statuscentral/models"
)

// Store is an interface that the storage implementers should implement
type Store interface {
	// Services
	CreateService(service *models.Service) error
	UpdateService(service *models.Service) error
	GetServices() ([]*models.Service, error)
	GetServicesEnabled() ([]*models.Service, error)
	GetServiceByName(name string) (*models.Service, error)
	GetServiceByID(id int) (*models.Service, error)

	// Regions
	CreateRegion(region *models.Region) error
	UpdateRegion(region *models.Region) error
	GetRegions() ([]*models.Region, error)
	GetRegionByCodeAndServiceName(regionCode, serviceName string) (*models.Region, error)
	DeleteRegion(id int) error

	// Incidents
	CreateIncident(incident *models.Incident) error
	UpdateIncident(incident *models.Incident) error
	GetIncidents(latest bool, pagination models.Pagination) ([]*models.Incident, error)
	GetIncidentByID(id int) (*models.Incident, error)
	DeleteIncident(id int) error

	// Incident Updates
	CreateIncidentUpdate(incidentID int, update *models.StatusUpdate) error
	GetIncidentUpdateByID(incidentID int, updateID int) (*models.StatusUpdate, error)
	GetIncidentUpdatesByIncidentID(incidentID int) ([]*models.StatusUpdate, error)
	DeleteIncidentUpdateByID(incidentID int, updateID int) error

	// Scheduled Maintenance
	CreateScheduledMaintenance(scheduledMaintenance *models.ScheduledMaintenance) error
	UpdateScheduledMaintenance(scheduledMaintenance *models.ScheduledMaintenance) error
	GetScheduledMaintenance(latest bool) ([]*models.ScheduledMaintenance, error)
	GetScheduledMaintenanceByID(id int) (*models.ScheduledMaintenance, error)
	DeleteScheduledMaintenance(id int) error

	// Scheduled Maintenance Updates
	CreateScheduledMaintenanceUpdate(maintenanceID int, update *models.StatusUpdate) error
	GetScheduledMaintenanceUpdateByID(maintenanceID int, updateID int) (*models.StatusUpdate, error)
	GetScheduledMaintenanceUpdatesByMaintenanceID(maintenanceID int) ([]*models.StatusUpdate, error)
	DeleteScheduledMaintenanceUpdateByID(maintenanceID int, updateID int) error

	CheckDb() error
	Snapshot(w io.Writer) error
}
