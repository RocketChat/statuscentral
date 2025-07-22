package core

import (
	"bytes"
	"errors"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/RocketChat/statuscentral/config"
	"github.com/RocketChat/statuscentral/models"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// GetScheduledMaintenance retrieves the scheduled maintenance from the storage layer
func GetScheduledMaintenance(latest bool) ([]*models.ScheduledMaintenance, error) {
	return _dataStore.GetScheduledMaintenance(latest)
}

// GetScheduledMaintenanceByID retrieves the scheduled maintenance by id, both scheduled maintenance and error will be nil if none found
func GetScheduledMaintenanceByID(id int) (*models.ScheduledMaintenance, error) {
	return _dataStore.GetScheduledMaintenanceByID(id)
}

// SendScheduledMaintenanceTwitter sends the info about the scheduled maintenance to the offical Rocket.Chat Cloud twitter account.
func SendScheduledMaintenanceTwitter(incident *models.ScheduledMaintenance) (int64, error) {
	conf := oauth1.NewConfig(config.Config.Twitter.ConsumerKey, config.Config.Twitter.ConsumerSecret)
	token := oauth1.NewToken(config.Config.Twitter.AccessToken, config.Config.Twitter.AccessSecret)
	http := conf.Client(oauth1.NoContext, token)
	http.Timeout = 5 * time.Second

	client := twitter.NewClient(http)
	tmpl, err := template.ParseFiles("templates/incident/tweet/maintenance.tmpl")
	if err != nil {
		return 0, err
	}

	b := &bytes.Buffer{}
	if err = tmpl.ExecuteTemplate(b, tmpl.Name(), incident); err != nil {
		return 0, err
	}

	tweet, _, err := client.Statuses.Update(b.String(), nil)
	if err != nil {
		return 0, err
	}

	return tweet.ID, nil
}

// SendScheduledMaintenanceUpdateTwitter sends the info about the update to scheduled maintenance to the twitter
func SendScheduledMaintenanceUpdateTwitter(scheduledMaintenance *models.ScheduledMaintenance, update *models.StatusUpdate) (int64, error) {
	conf := oauth1.NewConfig(config.Config.Twitter.ConsumerKey, config.Config.Twitter.ConsumerSecret)
	token := oauth1.NewToken(config.Config.Twitter.AccessToken, config.Config.Twitter.AccessSecret)
	http := conf.Client(oauth1.NoContext, token)
	http.Timeout = 5 * time.Second

	client := twitter.NewClient(http)
	tmpl, err := template.ParseFiles("templates/incident/tweet/maintenance.tmpl")
	if err != nil {
		return 0, err
	}

	b := &bytes.Buffer{}
	if err = tmpl.ExecuteTemplate(b, tmpl.Name(), scheduledMaintenance); err != nil {
		return 0, err
	}

	tweet, _, err := client.Statuses.Update(b.String(), nil)
	if err != nil {
		return 0, err
	}

	return tweet.ID, nil
}

// CreateScheduledMaintenance creates scheduled maintenance in the storage layer
func CreateScheduledMaintenance(scheduledMaintenance *models.ScheduledMaintenance) (*models.ScheduledMaintenance, error) {
	ensureScheduledMaintenanceDefaults(scheduledMaintenance)

	if len(scheduledMaintenance.Updates) > 0 {
		scheduledMaintenance.Updates = make([]*models.StatusUpdate, 0)
	}

	scheduledMaintenance.CreatedAt = time.Now()

	if scheduledMaintenance.PlannedStart.Before(scheduledMaintenance.CreatedAt) || scheduledMaintenance.PlannedEnd.Before(scheduledMaintenance.CreatedAt) {
		return nil, errors.New("start and end date must be in the future")
	}

	if err := _dataStore.CreateScheduledMaintenance(scheduledMaintenance); err != nil {
		return nil, err
	}

	// Todo: we need to figure out how we want this to look
	/*if config.Config.Twitter.Enabled {
		tweetID, err := SendScheduledMaintenanceTwitter(scheduledMaintenance)
		if err == nil {
			scheduledMaintenance.OriginalTweetID = tweetID
			if err := _dataStore.UpdateScheduledMaintenance(scheduledMaintenance); err != nil {
				return nil, err
			}
		}
	}*/

	return scheduledMaintenance, nil
}

// PatchScheduledMaintenance updates the scheduled maintenance in the storage layer
func PatchScheduledMaintenance(scheduledMaintenance *models.ScheduledMaintenance) error {

	existingMaintenance, err := _dataStore.GetScheduledMaintenanceByID(scheduledMaintenance.ID)
	if err != nil {
		return err
	}

	if existingMaintenance == nil {
		return errors.New("invalid scheduledMaintenance")
	}

	scheduledMaintenance.UpdatedAt = time.Now()

	if scheduledMaintenance.Title == "" {
		scheduledMaintenance.Title = existingMaintenance.Title
	}

	if scheduledMaintenance.Description == "" {
		scheduledMaintenance.Description = existingMaintenance.Description
	}

	if scheduledMaintenance.PlannedStart.Before(scheduledMaintenance.CreatedAt) || scheduledMaintenance.PlannedEnd.Before(scheduledMaintenance.CreatedAt) {
		return errors.New("start and end date must be in the future")
	}

	if scheduledMaintenance.PlannedStart.IsZero() {
		scheduledMaintenance.PlannedStart = existingMaintenance.PlannedStart
	}

	if scheduledMaintenance.PlannedEnd.IsZero() {
		scheduledMaintenance.PlannedEnd = existingMaintenance.PlannedEnd
	}

	scheduledMaintenance.Completed = existingMaintenance.Completed

	scheduledMaintenance.Updates = existingMaintenance.Updates
	scheduledMaintenance.CreatedAt = existingMaintenance.CreatedAt
	scheduledMaintenance.Services = existingMaintenance.Services
	scheduledMaintenance.LatestTweetID = existingMaintenance.LatestTweetID
	scheduledMaintenance.OriginalTweetID = existingMaintenance.OriginalTweetID

	if err := _dataStore.UpdateScheduledMaintenance(scheduledMaintenance); err != nil {
		return err
	}

	return nil
}

// DeleteScheduledMaintenance removes the scheduled maintenance from the storage layer
func DeleteScheduledMaintenance(id int) error {
	return _dataStore.DeleteScheduledMaintenance(id)
}

// CreateScheduledMaintenanceUpdate creates an update for a scheduled maintenance
func CreateScheduledMaintenanceUpdate(incidentID int, update *models.StatusUpdate) (*models.ScheduledMaintenance, error) {
	if incidentID <= 0 {
		return nil, errors.New("invalid incident id")
	}

	if update.Message == "" {
		return nil, errors.New("message property is missing")
	}

	if update.Status == "" {
		return nil, errors.New("status property is missing")
	}

	status, ok := models.IncidentStatuses[strings.ToLower(update.Status.String())]
	if !ok {
		return nil, errors.New("invalid status value")
	}

	update.Status = status

	if err := _dataStore.CreateScheduledMaintenanceUpdate(incidentID, update); err != nil {
		return nil, err
	}

	scheduledMaintenance, err := _dataStore.GetScheduledMaintenanceByID(incidentID)
	if err != nil {
		return nil, err
	}

	if status != models.IncidentStatusResolved {
		for _, s := range update.Services {
			if err := updateServiceToStatus(s.Name, s.Status); err != nil {
				return nil, err
			}

			// Update the status on the incident.  Makes it easier for those utilizing api
			for i, si := range scheduledMaintenance.Services {
				if s.Name == si.Name {
					scheduledMaintenance.Services[i].Status = s.Status
				}
			}

			for _, regionCode := range s.Regions {
				if err := updateRegionToStatus(regionCode, s.Name, s.Status); err != nil {
					return nil, err
				}
			}
		}
	} else {
		for i, s := range scheduledMaintenance.Services {
			scheduledMaintenance.Services[i].Status = models.ServiceStatusNominal
			if err := updateServiceToStatus(s.Name, models.ServiceStatusNominal); err != nil {
				return nil, err
			}

			for _, regionCode := range s.Regions {
				if err := updateRegionToStatus(regionCode, s.Name, models.ServiceStatusNominal); err != nil {
					return nil, err
				}
			}
		}

		scheduledMaintenance.Completed = true
	}

	if err := _dataStore.UpdateScheduledMaintenance(scheduledMaintenance); err != nil {
		return nil, err
	}

	if config.Config.Twitter.Enabled {
		tweetID, err := SendScheduledMaintenanceUpdateTwitter(scheduledMaintenance, update)
		if err == nil && tweetID != 0 {
			scheduledMaintenance.LatestTweetID = tweetID
			if err := _dataStore.UpdateScheduledMaintenance(scheduledMaintenance); err != nil {
				return nil, err
			}
		}
	}

	return scheduledMaintenance, nil
}

// GetScheduledMaintenanceUpdates gets updates for a scheduled maintenance
func GetScheduledMaintenanceUpdates(maintenanceID int) ([]*models.StatusUpdate, error) {
	if maintenanceID <= 0 {
		return nil, errors.New("invalid incident id")
	}

	updates, err := _dataStore.GetScheduledMaintenanceUpdatesByMaintenanceID(maintenanceID)
	if err != nil {
		return nil, errors.New("unable to get incident update")
	}

	return updates, nil
}

// GetScheduledMaintenanceUpdate gets an update for a scheduled maintenance
func GetScheduledMaintenanceUpdate(incidentID int, updateID int) (*models.StatusUpdate, error) {
	if incidentID <= 0 {
		return nil, errors.New("invalid incident id")
	}

	update, err := _dataStore.GetScheduledMaintenanceUpdateByID(incidentID, updateID)
	if err != nil {
		return nil, errors.New("unable to get incident update")
	}

	return update, nil
}

// DeleteScheduledMaintenanceUpdate deletes an update for a scheduled maintenance
func DeleteScheduledMaintenanceUpdate(incidentID int, updateID int) error {
	if incidentID <= 0 {
		return errors.New("invalid incident id")
	}

	update, err := _dataStore.GetScheduledMaintenanceUpdateByID(incidentID, updateID)
	if err != nil {
		return errors.New("unable to get incident update")
	}

	if update == nil {
		return nil
	}

	if err := _dataStore.DeleteScheduledMaintenanceUpdateByID(incidentID, updateID); err != nil {
		return err
	}

	return nil
}

func ensureScheduledMaintenanceDefaults(scheduledMaintenance *models.ScheduledMaintenance) {
	if scheduledMaintenance.Updates == nil {
		scheduledMaintenance.Updates = make([]*models.StatusUpdate, 0)
	}

	if scheduledMaintenance.Title == "" {
		scheduledMaintenance.Title = "Scheduled Maintenance"
	}
}

// AggregateScheduledMaintenance aggregates scheduled maintenance events by day.
// It only includes days that have at least one maintenance event.
func AggregateScheduledMaintenance(scheduledMaintenance []*models.ScheduledMaintenance) models.AggregatedScheduledMaintenances {
	// Use a map to group maintenance events by day efficiently.
	maintenanceByDay := make(map[time.Time][]*models.ScheduledMaintenance)
	for _, maintenance := range scheduledMaintenance {
		day := truncateToDay(maintenance.PlannedStart)
		maintenanceByDay[day] = append(maintenanceByDay[day], maintenance)
	}

	// Extract all unique days from the map keys into a slice for sorting.
	sortedDays := make([]time.Time, 0, len(maintenanceByDay))
	for day := range maintenanceByDay {
		sortedDays = append(sortedDays, day)
	}

	// Sort the days in reverse chronological order (newest first).
	sort.Slice(sortedDays, func(i, j int) bool {
		return sortedDays[i].After(sortedDays[j])
	})

	aggregatedScheduledMaintenances := models.AggregatedScheduledMaintenances{
		Days:  make([]models.ScheduledMaintenanceAggregatedByDay, 0, len(sortedDays)),
		Count: len(scheduledMaintenance),
	}

	// Build the final aggregated list from the sorted days.
	for _, day := range sortedDays {
		aggregatedScheduledMaintenances.Days = append(aggregatedScheduledMaintenances.Days, models.ScheduledMaintenanceAggregatedByDay{
			Time:                 day,
			ScheduledMaintenance: maintenanceByDay[day],
		})
	}

	return aggregatedScheduledMaintenances
}

func GetActiveMaintenance(scheduledMaintenances []*models.ScheduledMaintenance) *models.ScheduledMaintenance {
	for _, scheduledMaintenance := range scheduledMaintenances {
		if scheduledMaintenance.PlannedStart.After(time.Now()) && !scheduledMaintenance.Completed {
			return scheduledMaintenance
		}
	}

	return nil
}
