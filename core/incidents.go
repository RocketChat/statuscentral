package core

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/RocketChat/statuscentral/config"
	"github.com/RocketChat/statuscentral/models"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// GetIncidents retrieves the incidents from the storage layer
func GetIncidents(latest bool, pagination models.Pagination) ([]*models.Incident, error) {
	return _dataStore.GetIncidents(latest, pagination)
}

// GetIncidentByID retrieves the incident by id, both incident and error will be nil if none found
func GetIncidentByID(id int) (*models.Incident, error) {
	return _dataStore.GetIncidentByID(id)
}

// SendIncidentTwitter sends the incident info to the offical Rocket.Chat Cloud twitter account.
func SendIncidentTwitter(incident *models.Incident) (int64, error) {
	conf := oauth1.NewConfig(config.Config.Twitter.ConsumerKey, config.Config.Twitter.ConsumerSecret)
	token := oauth1.NewToken(config.Config.Twitter.AccessToken, config.Config.Twitter.AccessSecret)
	http := conf.Client(oauth1.NoContext, token)
	http.Timeout = 5 * time.Second

	client := twitter.NewClient(http)
	tmpl, err := template.ParseFiles("templates/incident/tweet/create.tmpl")
	if err != nil {
		return 0, err
	}

	b := &bytes.Buffer{}
	if err := tmpl.ExecuteTemplate(b, tmpl.Name(), incident); err != nil {
		return 0, err
	}

	tweet, _, err := client.Statuses.Update(b.String(), nil)
	if err != nil {
		return 0, err
	}

	return tweet.ID, nil
}

// SendIncidentUpdateTwitter sends the incident update info to the offical Rocket.Chat Cloud twitter account.
func SendIncidentUpdateTwitter(incident *models.Incident, update *models.StatusUpdate) (int64, error) {
	conf := oauth1.NewConfig(config.Config.Twitter.ConsumerKey, config.Config.Twitter.ConsumerSecret)
	token := oauth1.NewToken(config.Config.Twitter.AccessToken, config.Config.Twitter.AccessSecret)
	http := conf.Client(oauth1.NoContext, token)
	http.Timeout = 5 * time.Second

	client := twitter.NewClient(http)
	tmpl, err := template.ParseFiles("templates/incident/tweet/update.tmpl")
	if err != nil {
		return 0, err
	}

	b := &bytes.Buffer{}
	if err := tmpl.ExecuteTemplate(b, tmpl.Name(), map[string]interface{}{
		"update":   update,
		"incident": incident,
	}); err != nil {
		return 0, err
	}

	var params = &twitter.StatusUpdateParams{}
	if incident.OriginalTweetID != 0 {
		params.InReplyToStatusID = incident.OriginalTweetID
	}

	tweet, _, err := client.Statuses.Update(fmt.Sprintf("%s", b.String()), params)
	if err != nil {
		return 0, err
	}

	return tweet.ID, nil
}

// CreateIncident creates the incident in the storage layer
func CreateIncident(incident *models.Incident) (*models.Incident, error) {
	ensureIncidentDefaults(incident)

	if incident.Status == models.IncidentStatusScheduledMaintenance {
		incident.IsMaintenance = true
	}

	if len(incident.Updates) == 0 {
		if incident.IsMaintenance {
			update := models.StatusUpdate{
				Time:   incident.Time,
				Status: incident.Status,
				Message: fmt.Sprintf("Starts at %s with a scheduled end at %s",
					incident.Maintenance.Start.Format(time.RFC1123Z),
					incident.Maintenance.End.Format(time.RFC1123Z)),
			}

			incident.Updates = append(incident.Updates, &update)
		} else {
			update := models.StatusUpdate{
				Time:    incident.Time,
				Status:  incident.Status,
				Message: "Initial status of " + incident.Status.String(),
			}

			incident.Updates = append(incident.Updates, &update)
		}
	}

	if incident.IsMaintenance {
		for _, s := range incident.Services {

			if err := updateServiceToStatus(s.Name, models.ServiceStatusScheduledMaintenance); err != nil {
				return nil, err
			}

			for _, regionCode := range s.Regions {
				if err := updateRegionToStatus(regionCode, s.Name, models.ServiceStatusScheduledMaintenance); err != nil {
					return nil, err
				}
			}
		}

	} else {
		for _, s := range incident.Services {
			if err := updateServiceToStatus(s.Name, s.Status); err != nil {
				return nil, err
			}

			for _, regionCode := range s.Regions {
				if err := updateRegionToStatus(regionCode, s.Name, s.Status); err != nil {
					return nil, err
				}
			}
		}
	}

	if err := _dataStore.CreateIncident(incident); err != nil {
		return nil, err
	}

	if config.Config.Twitter.Enabled {
		tweetID, err := SendIncidentTwitter(incident)
		if err == nil {
			incident.OriginalTweetID = tweetID
			if err := _dataStore.UpdateIncident(incident); err != nil {
				return nil, err
			}
		}
	}

	return incident, nil
}

// DeleteIncident removes the incident from the storage layer
func DeleteIncident(id int) error {
	return _dataStore.DeleteIncident(id)
}

// CreateIncidentUpdate creates an update for an incident
func CreateIncidentUpdate(incidentID int, update *models.StatusUpdate) (*models.Incident, error) {
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

	if err := _dataStore.CreateIncidentUpdate(incidentID, update); err != nil {
		return nil, err
	}

	incident, err := _dataStore.GetIncidentByID(incidentID)
	if err != nil {
		return nil, err
	}

	if status != models.IncidentStatusResolved {
		for _, s := range update.Services {
			if err := updateServiceToStatus(s.Name, s.Status); err != nil {
				return nil, err
			}

			// Update the status on the incident.  Makes it easier for those utilizing api
			for i, si := range incident.Services {
				if s.Name == si.Name {
					incident.Services[i].Status = s.Status
				}
			}

			for _, regionCode := range s.Regions {
				if err := updateRegionToStatus(regionCode, s.Name, s.Status); err != nil {
					return nil, err
				}
			}
		}
	} else {
		for _, s := range incident.Services {
			s.Status = models.ServiceStatusNominal

			if err := updateServiceToStatus(s.Name, models.ServiceStatusNominal); err != nil {
				return nil, err
			}

			for _, regionCode := range s.Regions {
				if err := updateRegionToStatus(regionCode, s.Name, models.ServiceStatusNominal); err != nil {
					return nil, err
				}
			}
		}
	}

	if err := _dataStore.UpdateIncident(incident); err != nil {
		return nil, err
	}

	if config.Config.Twitter.Enabled {
		tweetID, err := SendIncidentUpdateTwitter(incident, update)
		if err == nil && tweetID != 0 {
			incident.LatestTweetID = tweetID
			if err := _dataStore.UpdateIncident(incident); err != nil {
				return nil, err
			}
		}
	}

	return incident, nil
}

// GetIncidentUpdates gets updates for an incident
func GetIncidentUpdates(incidentID int) ([]*models.StatusUpdate, error) {
	if incidentID <= 0 {
		return nil, errors.New("invalid incident id")
	}

	updates, err := _dataStore.GetIncidentUpdatesByIncidentID(incidentID)
	if err != nil {
		return nil, errors.New("unable to get incident update")
	}

	return updates, nil
}

// GetIncidentUpdate gets an update for an incident
func GetIncidentUpdate(incidentID int, updateID int) (*models.StatusUpdate, error) {
	if incidentID <= 0 {
		return nil, errors.New("invalid incident id")
	}

	update, err := _dataStore.GetIncidentUpdateByID(incidentID, updateID)
	if err != nil {
		return nil, errors.New("unable to get incident update")
	}

	return update, nil
}

// DeleteIncidentUpdate deletes an update for an incident
func DeleteIncidentUpdate(incidentID int, updateID int) error {
	if incidentID <= 0 {
		return errors.New("invalid incident id")
	}

	update, err := _dataStore.GetIncidentUpdateByID(incidentID, updateID)
	if err != nil {
		return errors.New("unable to get incident update")
	}

	if update == nil {
		return nil
	}

	if err := _dataStore.DeleteIncidentUpdateByID(incidentID, updateID); err != nil {
		return err
	}

	return nil
}

func ensureIncidentDefaults(incident *models.Incident) {
	if incident.Updates == nil {
		incident.Updates = make([]*models.StatusUpdate, 0)
	}

	if incident.Status == "" {
		incident.Status = models.IncidentDefaultStatus
	}

	if incident.Title == "" {
		incident.Title = "Unknown"
	}

	if incident.Time.IsZero() {
		incident.Time = time.Now()
	}
}

// truncateToDay is a helper function that normalizes a time.Time object
// to the beginning of its calendar day (00:00:00).
func truncateToDay(t time.Time) time.Time {
	// Preserves the location (timezone) of the original time.
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func AggregateIncidents(incidents []*models.Incident, showEmptyDays bool) models.AggregatedIncidents {
	incidentsByDay := make(map[time.Time][]*models.Incident)

	// If showing empty days, "prime" the map with empty slices for recent days.
	if showEmptyDays {
		now := time.Now()
		for i := 0; i < config.Config.Website.EmptyDaysToShow; i++ {
			day := truncateToDay(now.AddDate(0, 0, -i))
			incidentsByDay[day] = []*models.Incident{}
		}
	}

	// Group the actual incidents. This will append to existing empty slices for recent days
	// or create new map entries for older days that have incidents.
	for _, incident := range incidents {
		day := truncateToDay(incident.Time)
		incidentsByDay[day] = append(incidentsByDay[day], incident)
	}

	// Extract all unique days from the map keys into a slice for sorting.
	sortedDays := make([]time.Time, 0, len(incidentsByDay))
	for day := range incidentsByDay {
		sortedDays = append(sortedDays, day)
	}

	// Sort the days in reverse chronological order (newest first).
	sort.Slice(sortedDays, func(i, j int) bool {
		return sortedDays[i].After(sortedDays[j])
	})

	// Build the final aggregated list from the sorted days.
	aggregatedIncidents := make(models.AggregatedIncidents, 0, len(sortedDays))
	for _, day := range sortedDays {
		aggregatedIncidents = append(aggregatedIncidents, models.AggregatedIncident{
			Time:      day,
			Incidents: incidentsByDay[day],
		})
	}

	return aggregatedIncidents
}
