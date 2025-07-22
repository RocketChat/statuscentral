package v1

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/RocketChat/statuscentral/core"
	"github.com/RocketChat/statuscentral/models"
	"github.com/gin-gonic/gin"
)

// IncidentsGetAll gets all of the incidents, latest depends on the "?all=true" query
// @Summary Gets list of incidents
// @ID incidents-getall
// @Tags incident
// @Produce json
// @Success 200 {object} []models.Incident
// @Router /v1/incidents [get]
func IncidentsGetAll(c *gin.Context) {
	allParam := c.Query("all")

	latest := true
	if allParam != "" && allParam == "true" {
		latest = false
	}

	incidents, err := core.GetIncidents(latest, models.Pagination{})
	if err != nil {
		internalErrorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, incidents)
}

// IncidentGetOne gets one incident by the provided id
// @Summary Gets one incident
// @ID incidents-getOne
// @Tags incident
// @Produce json
// @Success 200 {object} models.Incident
// @Router /v1/incidents/{id} [get]
func IncidentGetOne(c *gin.Context) {
	idParam := c.Param("id")

	if idParam == "" {
		badRequestHandlerDetailed(c, errors.New("invalid incident id passed"))
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		badRequestHandlerDetailed(c, err)
		return
	}

	incident, err := core.GetIncidentByID(id)
	if err != nil {
		internalErrorHandlerDetailed(c, err)
		return
	}

	if incident == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, incident)
}

// IncidentCreate creates the incident, ensuring the database is correct
// @Summary Creates a new incident
// @ID incident-create
// @Tags incident
// @Accept json
// @Param region body models.Incident true "Incident object"
// @Produce json
// @Success 200 {object} models.Incident
// @Router /v1/incidents [post]
func IncidentCreate(c *gin.Context) {
	var incident models.Incident

	if err := c.BindJSON(&incident); err != nil {
		return
	}

	if incident.Title == "" {
		badRequestHandlerDetailed(c, errors.New("title must be provided"))
		return
	}

	if incident.Status == models.IncidentStatusScheduledMaintenance {
		badRequestHandlerDetailed(c, errors.New("use scheduled maintenance endpoints to schedule maintenance"))
		return
	}

	inc, err := core.CreateIncident(&incident)
	if err != nil {
		internalErrorHandlerDetailed(c, err)
		return
	}

	c.JSON(http.StatusCreated, &inc)
}

// IncidentDelete removes the service, ensuring the database is correct
// @Summary Deletes an incidents
// @ID incidents-delete
// @Tags incident
// @Produce json
// @Success 200 {object} []models.Incident
// @Router /v1/incidents/{id} [delete]
func IncidentDelete(c *gin.Context) {
	idParam := c.Param("id")

	if idParam == "" {
		badRequestHandlerDetailed(c, errors.New("invalid incident id passed"))
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		badRequestHandlerDetailed(c, err)
		return
	}

	if err := core.DeleteIncident(id); err != nil {
		internalErrorHandlerDetailed(c, err)
		return
	}

	c.Status(http.StatusOK)
}

// IncidentUpdateCreate creates an update for an incident
// @Summary Creates a new incident update
// @ID incident-create-update
// @Tags incident
// @Accept json
// @Param region body models.IncidentUpdate true "Incident update object"
// @Param id path integer true "Incident id"
// @Produce json
// @Success 200 {object} models.IncidentUpdate
// @Router /v1/incidents/{id}/updates [post]
func IncidentUpdateCreate(c *gin.Context) {
	idParam := c.Param("id")

	if idParam == "" {
		badRequestHandlerDetailed(c, errors.New("invalid incident id passed"))
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		badRequestHandlerDetailed(c, err)
		return
	}

	var update models.StatusUpdate
	if err := c.BindJSON(&update); err != nil {
		return
	}

	if update.Message == "" {
		badRequestHandlerDetailed(c, errors.New("message is missing"))
		return
	}

	if update.Status == "" {
		badRequestHandlerDetailed(c, errors.New("status is missing"))
		return
	}

	status, ok := models.IncidentStatuses[strings.ToLower(update.Status.String())]
	if !ok {
		badRequestHandlerDetailed(c, errors.New("invalid status value"))
		return
	}

	update.Status = status

	incident, err := core.CreateIncidentUpdate(id, &update)
	if err != nil {
		internalErrorHandlerDetailed(c, err)
		return
	}

	c.JSON(http.StatusCreated, incident)
}

// IncidentUpdatesGetAll gets updates for an incident
// @Summary Gets incident updates
// @ID incident-update-getall
// @Tags incident-update
// @Produce json
// @Success 200 {object} []models.IncidentUpdate
// @Router /v1/incidents/{id}/updates [get]
func IncidentUpdatesGetAll(c *gin.Context) {
	idParam := c.Param("id")

	if idParam == "" {
		badRequestHandlerDetailed(c, errors.New("invalid incident id passed"))
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		badRequestHandlerDetailed(c, err)
		return
	}

	updates, err := core.GetIncidentUpdates(id)
	if err != nil {
		internalErrorHandlerDetailed(c, err)
		return
	}

	c.JSON(http.StatusOK, updates)
}

// IncidentUpdateGetOne gets an update for an incident
// @Summary Gets one incident update
// @ID incident-update-getone
// @Tags incident-update
// @Produce json
// @Success 200 {object} models.IncidentUpdate
// @Router /v1/incidents/{id}/updates/{updateId} [get]
func IncidentUpdateGetOne(c *gin.Context) {
	idParam := c.Param("id")
	updateIdParam := c.Param("updateId")

	if idParam == "" {
		badRequestHandlerDetailed(c, errors.New("invalid incident id passed"))
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		badRequestHandlerDetailed(c, err)
		return
	}

	if updateIdParam == "" {
		badRequestHandlerDetailed(c, errors.New("invalid update id passed"))
		return
	}

	updateID, err := strconv.Atoi(updateIdParam)
	if err != nil {
		badRequestHandlerDetailed(c, err)
		return
	}

	update, err := core.GetIncidentUpdate(id, updateID)
	if err != nil {
		internalErrorHandlerDetailed(c, err)
		return
	}

	c.JSON(http.StatusOK, update)
}

// IncidentUpdateDelete deletes an update for an incident
// @Summary Deletes one incident update
// @ID incident-update-delete
// @Tags incident-update
// @Produce json
// @Success 200 {object} models.IncidentUpdate
// @Router /v1/incidents/{id}/updates/{updateId} [delete]
func IncidentUpdateDelete(c *gin.Context) {
	idParam := c.Param("id")
	updateIdParam := c.Param("updateId")

	if idParam == "" {
		badRequestHandlerDetailed(c, errors.New("invalid incident id passed"))
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		badRequestHandlerDetailed(c, err)
		return
	}

	if updateIdParam == "" {
		badRequestHandlerDetailed(c, errors.New("invalid update id passed"))
		return
	}

	updateId, err := strconv.Atoi(updateIdParam)
	if err != nil {
		badRequestHandlerDetailed(c, err)
		return
	}

	if err := core.DeleteIncidentUpdate(id, updateId); err != nil {
		internalErrorHandlerDetailed(c, err)
		return
	}

	c.Status(http.StatusOK)
}
