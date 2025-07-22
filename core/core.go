package core

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/RocketChat/statuscentral/models"
	"github.com/RocketChat/statuscentral/store"
	"github.com/RocketChat/statuscentral/store/boltstore"
)

var _dataStore store.Store

// TwistItUp takes everything and starts the core up
func TwistItUp() error {
	store, err := boltstore.New()
	if err != nil {
		log.Fatalln(err)
	}

	_dataStore = store

	// Now that we have a store, let's ensure the services and regions from the config exist
	if err := createServicesFromConfig(); err != nil {
		log.Fatalln(err)
	}

	// Regions always need to be created AFTER services due to regions being tethered to services
	if err := createRegionsFromConfig(); err != nil {
		log.Fatalln(err)
	}

	return nil
}

func RunMigrations() error {
	log.Println("Run migrations")
	// Migrate incidents over to scheduled maintenances
	incidents, err := _dataStore.GetIncidents(false, models.Pagination{})
	if err != nil {
		return err
	}

	for _, incident := range incidents {
		log.Println(fmt.Sprintf("%d - %s", incident.ID, incident.Title))
		if incident.IsMaintenance {
			scheduledMaintenance := models.ScheduledMaintenance{
				Title:           "Scheduled Maintenance",
				Description:     incident.Title,
				Services:        incident.Services,
				OriginalTweetID: incident.OriginalTweetID,
				LatestTweetID:   incident.LatestTweetID,
				PlannedStart:    incident.Maintenance.Start,
				PlannedEnd:      incident.Maintenance.End,
				CreatedAt:       incident.Time,
				UpdatedAt:       incident.UpdatedAt,
			}

			if incident.Maintenance.End.Before(time.Now()) {
				scheduledMaintenance.Completed = true
			}

			for _, update := range scheduledMaintenance.Updates {
				if update.Status == models.IncidentStatusScheduledMaintenance {
					scheduledMaintenance.Description = update.Message
					continue
				}

				if update.Status == models.IncidentStatusResolved {
					scheduledMaintenance.Completed = true
				}

				scheduledMaintenance.Updates = append(scheduledMaintenance.Updates, update)
			}

			if err := _dataStore.CreateScheduledMaintenance(&scheduledMaintenance); err != nil {
				return err
			}

			if err := _dataStore.DeleteIncident(incident.ID); err != nil {
				return err
			}

			log.Println(fmt.Sprintf("Migrated %d - %s", incident.ID, incident.Title))
		}
	}

	return nil
}

// LivenessCheck checks the database to see if it responds to a ping
func LivenessCheck() error {
	return _dataStore.CheckDb()
}

// DBSnapshot returns snapshot of database
func DBSnapshot(w io.Writer) error {
	return _dataStore.Snapshot(w)
}
