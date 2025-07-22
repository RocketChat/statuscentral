package boltstore

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/RocketChat/statuscentral/config"
	"github.com/RocketChat/statuscentral/models"
	bolt "github.com/etcd-io/bbolt"
)

func (s *boltStore) GetScheduledMaintenance(latestOnly bool) ([]*models.ScheduledMaintenance, error) {
	tx, err := s.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	cursor := tx.Bucket(scheduledMaintenanceBucket).Cursor()

	days := config.Config.Website.EmptyDaysToShow
	to := time.Now()
	from := to.Add(time.Duration(-days*24) * time.Hour).Truncate(24 * time.Hour)

	scheduledMaintenances := make([]*models.ScheduledMaintenance, 0)
	for k, data := cursor.First(); k != nil; k, data = cursor.Next() {
		var m models.ScheduledMaintenance
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}

		if latestOnly {
			if m.CreatedAt.Before(to) && m.CreatedAt.After(from) {
				scheduledMaintenances = append(scheduledMaintenances, &m)
				continue
			}
		} else {
			scheduledMaintenances = append(scheduledMaintenances, &m)
		}

	}

	return scheduledMaintenances, nil
}

func (s *boltStore) GetScheduledMaintenanceByID(id int) (*models.ScheduledMaintenance, error) {
	tx, err := s.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bytes := tx.Bucket(scheduledMaintenanceBucket).Get(itob(id))
	if bytes == nil {
		return nil, nil
	}

	var m models.ScheduledMaintenance
	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (s *boltStore) CreateScheduledMaintenance(maintenance *models.ScheduledMaintenance) error {
	tx, err := s.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(scheduledMaintenanceBucket)

	seq, _ := bucket.NextSequence()
	maintenance.ID = int(seq)

	if maintenance.CreatedAt.IsZero() {
		maintenance.CreatedAt = time.Now()
	}

	maintenance.UpdatedAt = time.Now()

	buf, err := json.Marshal(maintenance)
	if err != nil {
		return err
	}

	if err := bucket.Put(itob(maintenance.ID), buf); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *boltStore) UpdateScheduledMaintenance(maintenance *models.ScheduledMaintenance) error {
	if maintenance.ID <= 0 {
		return errors.New("invalid maintenance id")
	}

	tx, err := s.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(scheduledMaintenanceBucket)

	maintenance.UpdatedAt = time.Now()

	buf, err := json.Marshal(maintenance)
	if err != nil {
		return err
	}

	if err := bucket.Put(itob(maintenance.ID), buf); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *boltStore) DeleteScheduledMaintenance(id int) error {
	return s.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(scheduledMaintenanceBucket).Delete(itob(id))
	})
}

func (s *boltStore) CreateScheduledMaintenanceUpdate(maintenanceID int, update *models.StatusUpdate) error {
	tx, err := s.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(scheduledMaintenanceBucket)

	bytes := bucket.Get(itob(maintenanceID))
	if bytes == nil {
		return errors.New("no scheduled maintenance found by that id")
	}

	var i models.ScheduledMaintenance
	if err := json.Unmarshal(bytes, &i); err != nil {
		return err
	}

	// If none index is 0 and then len should always put +1
	nextUpdateID := len(i.Updates)

	update.ID = nextUpdateID

	if update.Time.IsZero() {
		update.Time = time.Now()
	}

	//i.Status = update.Status
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

func (s *boltStore) GetScheduledMaintenanceUpdateByID(maintenanceID int, updateId int) (*models.StatusUpdate, error) {
	tx, err := s.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bytes := tx.Bucket(scheduledMaintenanceBucket).Get(itob(maintenanceID))
	if bytes == nil {
		return nil, nil
	}

	var scheduledMaintenance models.ScheduledMaintenance
	if err := json.Unmarshal(bytes, &scheduledMaintenance); err != nil {
		return nil, err
	}

	for i, update := range scheduledMaintenance.Updates {
		if update.ID == updateId {
			return scheduledMaintenance.Updates[i], nil
		}
	}

	return nil, nil
}

func (s *boltStore) GetScheduledMaintenanceUpdatesByMaintenanceID(maintenanceID int) ([]*models.StatusUpdate, error) {
	tx, err := s.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bytes := tx.Bucket(scheduledMaintenanceBucket).Get(itob(maintenanceID))
	if bytes == nil {
		return nil, nil
	}

	var scheduledMaintenance models.ScheduledMaintenance
	if err := json.Unmarshal(bytes, &scheduledMaintenance); err != nil {
		return nil, err
	}

	return scheduledMaintenance.Updates, nil
}

func (s *boltStore) DeleteScheduledMaintenanceUpdateByID(maintenanceID int, updateId int) error {
	tx, err := s.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bucket := tx.Bucket(scheduledMaintenanceBucket)

	bytes := bucket.Get(itob(maintenanceID))
	if bytes == nil {
		return nil
	}

	var scheduledMaintenance models.ScheduledMaintenance
	if err := json.Unmarshal(bytes, &scheduledMaintenance); err != nil {
		return err
	}

	updates := []*models.StatusUpdate{}

	for _, update := range scheduledMaintenance.Updates {
		if update.ID != updateId {
			updates = append(updates, update)
		}
	}

	scheduledMaintenance.Updates = updates

	scheduledMaintenance.UpdatedAt = time.Now()

	buf, err := json.Marshal(scheduledMaintenance)
	if err != nil {
		return err
	}

	if err := bucket.Put(itob(scheduledMaintenance.ID), buf); err != nil {
		return err
	}

	return tx.Commit()
}
