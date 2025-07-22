package boltstore

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/RocketChat/statuscentral/config"
	"github.com/RocketChat/statuscentral/models"
	bolt "github.com/etcd-io/bbolt"
)

// GetIncidents retrieves a paginated list of incidents.
// Incidents are returned from newest to oldest.
func (s *boltStore) GetIncidents(latestOnly bool, pagination models.Pagination) ([]*models.Incident, error) {
	// Provide sane defaults for pagination to prevent errors or fetching unlimited data.
	// Assuming models.Pagination has Limit and Offset fields.
	if pagination.Limit <= 0 {
		pagination.Limit = 25 // Default page size
	}

	if pagination.Offset < 0 {
		pagination.Offset = 0
	}

	if pagination.Limit > 50 {
		pagination.Limit = 50
	}

	tx, err := s.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	cursor := tx.Bucket(incidentBucket).Cursor()

	// Define the time range for the `latestOnly` filter.
	var from, to time.Time
	if latestOnly {
		days := config.Config.Website.DaysToAggregate
		to = time.Now()
		from = to.Add(time.Duration(-days*24) * time.Hour).Truncate(24 * time.Hour)
	}

	// Pre-allocate the slice with the required capacity.
	incidents := make([]*models.Incident, 0, pagination.Limit)
	skippedCount := 0

	// Iterate from the last (newest) key to the first (oldest).
	// This is more efficient and user-friendly for pagination than loading all records.
	for k, data := cursor.Last(); k != nil; k, data = cursor.Prev() {
		var i models.Incident
		if err := json.Unmarshal(data, &i); err != nil {
			// Depending on requirements, you might want to log this error and continue
			// instead of failing the entire request.
			return nil, err
		}

		// If latestOnly is true, apply the time filter.
		if latestOnly {
			if i.Time.Before(from) || i.Time.After(to) {
				continue // Skip incidents that are outside the desired time range.
			}
		}

		// This section handles the pagination offset. We skip the number of incidents
		// specified by the offset before we start collecting them for the current page.
		if skippedCount < pagination.Offset {
			skippedCount++
			continue
		}

		// Only add incidents to the slice if we haven't reached the page limit yet.
		if len(incidents) < pagination.Limit {
			incidents = append(incidents, &i)
		}
	}

	return incidents, nil
}

func (s *boltStore) GetIncidentByID(id int) (*models.Incident, error) {
	tx, err := s.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bytes := tx.Bucket(incidentBucket).Get(itob(id))
	if bytes == nil {
		return nil, nil
	}

	var i models.Incident
	if err := json.Unmarshal(bytes, &i); err != nil {
		return nil, err
	}

	return &i, nil
}

func (s *boltStore) CreateIncident(incident *models.Incident) error {
	tx, err := s.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(incidentBucket)

	seq, _ := bucket.NextSequence()
	incident.ID = int(seq)
	incident.UpdatedAt = time.Now()

	buf, err := json.Marshal(incident)
	if err != nil {
		return err
	}

	if err := bucket.Put(itob(incident.ID), buf); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *boltStore) UpdateIncident(incident *models.Incident) error {
	if incident.ID <= 0 {
		return errors.New("invalid incident id")
	}

	tx, err := s.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(incidentBucket)

	incident.UpdatedAt = time.Now()

	buf, err := json.Marshal(incident)
	if err != nil {
		return err
	}

	if err := bucket.Put(itob(incident.ID), buf); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *boltStore) DeleteIncident(id int) error {
	return s.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(incidentBucket).Delete(itob(id))
	})
}

func (s *boltStore) CreateIncidentUpdate(incidentID int, update *models.StatusUpdate) error {
	tx, err := s.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(incidentBucket)

	bytes := bucket.Get(itob(incidentID))
	if bytes == nil {
		return errors.New("no incident found by that id")
	}

	var i models.Incident
	if err := json.Unmarshal(bytes, &i); err != nil {
		return err
	}

	// If none index is 0 and then len should always put +1
	nextUpdateID := len(i.Updates)

	update.ID = nextUpdateID

	if update.Time.IsZero() {
		update.Time = time.Now()
	}

	i.Status = update.Status
	i.Updates = append(i.Updates, update)
	i.UpdatedAt = time.Now()

	buf, err := json.Marshal(i)
	if err != nil {
		return err
	}

	if err := bucket.Put(itob(i.ID), buf); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *boltStore) GetIncidentUpdateByID(incidentId int, updateId int) (*models.StatusUpdate, error) {
	tx, err := s.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bytes := tx.Bucket(incidentBucket).Get(itob(incidentId))
	if bytes == nil {
		return nil, nil
	}

	var incident models.Incident
	if err := json.Unmarshal(bytes, &incident); err != nil {
		return nil, err
	}

	for i, update := range incident.Updates {
		if update.ID == updateId {
			return incident.Updates[i], nil
		}
	}

	return nil, nil
}

func (s *boltStore) GetIncidentUpdatesByIncidentID(incidentId int) ([]*models.StatusUpdate, error) {
	tx, err := s.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bytes := tx.Bucket(incidentBucket).Get(itob(incidentId))
	if bytes == nil {
		return nil, nil
	}

	var incident models.Incident
	if err := json.Unmarshal(bytes, &incident); err != nil {
		return nil, err
	}

	return incident.Updates, nil
}

func (s *boltStore) DeleteIncidentUpdateByID(incidentId int, updateId int) error {
	tx, err := s.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(incidentBucket)

	bytes := bucket.Get(itob(incidentId))
	if bytes == nil {
		return nil
	}

	var incident models.Incident
	if err := json.Unmarshal(bytes, &incident); err != nil {
		return err
	}

	updates := []*models.StatusUpdate{}

	for _, update := range incident.Updates {
		if update.ID != updateId {
			updates = append(updates, update)
		}
	}

	incident.Updates = updates

	incident.UpdatedAt = time.Now()

	buf, err := json.Marshal(incident)
	if err != nil {
		return err
	}

	if err := bucket.Put(itob(incident.ID), buf); err != nil {
		return err
	}

	return tx.Commit()
}
