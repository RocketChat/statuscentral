package client

import (
	"fmt"

	"github.com/RocketChat/statuscentral/models"
)

// IncidentsInterface incidents interface
type IncidentsInterface interface {
	Create(incident *models.Incident) (returnedIncident *models.Incident, err error)
	Get(id int) (incident *models.Incident, err error)
	GetMultiple(latestOnly bool) (result []*models.Incident, err error)
	CreateStatusUpdate(incidentID int, statusUpdate *models.StatusUpdate) (returnedIncident *models.Incident, err error)
	Delete(incidentID int) error
	Patch(incidentID int, patchData interface{}) (returnedIncident *models.Incident, err error)
}

type incidents struct {
	client *Client
}

// Get Gets incident by id
func (i *incidents) Get(id int) (incident *models.Incident, err error) {

	req, err := i.client.buildRequest("GET", fmt.Sprintf("/api/v1/incidents/%d", id), nil)
	if err != nil {
		return nil, err
	}

	incident = &models.Incident{}

	resp, err := i.client.do(req, incident)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return incident, nil
}

func (i *incidents) GetMultiple(latestOnly bool) (result []*models.Incident, err error) {
	req, err := i.client.buildRequest(
		"GET",
		fmt.Sprintf(
			"/api/v1/incidents?all=%v",
			!latestOnly,
		),
		nil,
	)

	if err != nil {
		return nil, err
	}

	result = []*models.Incident{}

	resp, err := i.client.do(req, &result)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Create creates incident
func (i *incidents) Create(incident *models.Incident) (returnedIncident *models.Incident, err error) {
	req, err := i.client.buildRequest("POST", "/api/v1/incidents", incident)
	if err != nil {
		return nil, err
	}

	returnedIncident = &models.Incident{}

	resp, err := i.client.do(req, returnedIncident)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return returnedIncident, nil
}

// CreateStatusUpdate Creates a status update for an incident
func (i *incidents) CreateStatusUpdate(incidentID int, statusUpdate *models.StatusUpdate) (returnedIncident *models.Incident, err error) {
	req, err := i.client.buildRequest("POST", fmt.Sprintf("/api/v1/incidents/%d/updates", incidentID), statusUpdate)
	if err != nil {
		return nil, err
	}

	returnedIncident = &models.Incident{}

	resp, err := i.client.do(req, returnedIncident)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return returnedIncident, nil
}

// Delete deletes an incident
func (i *incidents) Delete(incidentID int) error {
	req, err := i.client.buildRequest(
		"DELETE",
		fmt.Sprintf("/api/v1/incidents/%d", incidentID),
		nil,
	)

	if err != nil {
		return err
	}

	resp, err := i.client.do(req, nil)
	if err != nil {
		return err
	}

	err = resp.Body.Close()

	return err
}

// Patch patches an incident
func (i *incidents) Patch(incidentID int, patchData interface{}) (returnedIncident *models.Incident, err error) {
	req, err := i.client.buildRequest("PATCH", fmt.Sprintf("/api/v1/incidents/%d", incidentID), patchData)
	if err != nil {
		return nil, err
	}

	returnedIncident = &models.Incident{}

	resp, err := i.client.do(req, returnedIncident)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return returnedIncident, nil
}
