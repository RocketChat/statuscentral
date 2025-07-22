package v1

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/RocketChat/statuscentral/models"

	"github.com/RocketChat/statuscentral/config"
	"github.com/RocketChat/statuscentral/core"
	"github.com/gin-gonic/gin"
)

// IndexHandler is the html controller for sending the html dashboard
func IndexHandler(c *gin.Context) {
	services, err := core.GetServicesEnabled()
	if err != nil {
		log.Println("Error while getting the services:")
		log.Println(err)
		handleIndexPageLoadingFromConfig(c)
		return
	}

	incidents, err := core.GetIncidents(true, models.Pagination{Limit: 30})
	if err != nil {
		log.Println("Error while getting the incidents:")
		log.Println(err)
		handleIndexPageLoadingFromConfig(c)
		return
	}

	scheduledMaintenance, err := core.GetScheduledMaintenance(true)
	if err != nil {
		log.Println("Error while getting the scheduled maintenance:")
		log.Println(err)
		handleIndexPageLoadingFromConfig(c)
		return
	}

	regions, err := core.GetRegions()
	if err != nil {
		log.Println("Error while getting the regions:")
		log.Println(err)
		handleIndexPageLoadingFromConfig(c)
		return
	}

	for _, service := range services {
		for _, region := range regions {
			if region.ServiceID == service.ID {
				service.Regions = append(service.Regions, *region)
			}
		}
	}

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"owner":                config.Config.Website.Title,
		"backgroundColor":      config.Config.Website.HeaderBgColor,
		"cacheBreaker":         config.Config.Website.CacheBreaker,
		"logo":                 "static/img/logo.svg",
		"services":             services,
		"mostCriticalStatus":   core.MostCriticalServiceStatus(services, regions),
		"incidents":            core.AggregateIncidents(incidents, true),
		"scheduledMaintenance": core.AggregateScheduledMaintenance(scheduledMaintenance),
	})
}

func handleIndexPageLoadingFromConfig(c *gin.Context) {
	services := make([]*models.Service, 0)
	for _, s := range config.Config.Services {
		service := &models.Service{
			Name:        s.Name,
			Description: s.Description,
			Status:      models.ServiceStatusUnknown,
		}

		services = append(services, service)
	}

	regions := make([]*models.Region, 0)
	for _, s := range config.Config.Regions {
		region := &models.Region{
			Name:        s.Name,
			Description: s.Description,
			Status:      models.ServiceStatusUnknown,
		}

		regions = append(regions, region)
	}

	for _, service := range services {
		for _, region := range regions {
			if region.ServiceID == service.ID {
				service.Regions = append(service.Regions, *region)
			}
		}
	}

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"owner":                config.Config.Website.Title,
		"backgroundColor":      config.Config.Website.HeaderBgColor,
		"cacheBreaker":         config.Config.Website.CacheBreaker,
		"logo":                 "static/img/logo.svg",
		"services":             services,
		"mostCriticalStatus":   models.ServiceStatusValues["Unknown"],
		"incidents":            core.AggregateIncidents(make([]*models.Incident, 0), true),
		"scheduledMaintenance": core.AggregateScheduledMaintenance(make([]*models.ScheduledMaintenance, 0)),
	})
}

func IncidentShortRedirectHandler(c *gin.Context) {
	if c.Param("id") == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	c.Redirect(http.StatusPermanentRedirect, fmt.Sprintf("/incidents/%s", c.Param("id")))
}

// IncidentDetailHandler is the html controller for displaying the incident details
func IncidentDetailHandler(c *gin.Context) {
	if c.Param("id") == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		internalErrorHandler(c, err)
		return
	}

	services, err := core.GetServices()
	if err != nil {
		log.Println("Error while getting the services:")
		log.Println(err)
		handleIndexPageLoadingFromConfig(c)
		return
	}

	regions, err := core.GetRegions()
	if err != nil {
		log.Println("Error while getting the regions:")
		log.Println(err)
		handleIndexPageLoadingFromConfig(c)
		return
	}

	for _, service := range services {
		for _, region := range regions {
			if region.ServiceID == service.ID {
				service.Regions = append(service.Regions, *region)
			}
		}
	}

	incident, err := core.GetIncidentByID(id)
	if err != nil {
		internalErrorHandler(c, err)
		return
	}

	if incident == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	c.HTML(http.StatusOK, "incidentDetail.tmpl", gin.H{
		"owner":              config.Config.Website.Title,
		"backgroundColor":    config.Config.Website.HeaderBgColor,
		"cacheBreaker":       config.Config.Website.CacheBreaker,
		"logo":               "static/img/logo.svg",
		"mostCriticalStatus": core.MostCriticalServiceStatus(services, regions),
		"services":           services,
		"incident":           incident,
	})
}

func ScheduledMaintenanceShortRedirectHandler(c *gin.Context) {
	if c.Param("id") == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	c.Redirect(http.StatusPermanentRedirect, fmt.Sprintf("/scheduled-maintenance/%s", c.Param("id")))
}

// ScheduledMaintenanceDetailHandler is the html controller for displaying the scheduled maintenance details
func ScheduledMaintenanceDetailHandler(c *gin.Context) {
	if c.Param("id") == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		internalErrorHandler(c, err)
		return
	}

	services, err := core.GetServices()
	if err != nil {
		log.Println("Error while getting the services:")
		log.Println(err)
		handleIndexPageLoadingFromConfig(c)
		return
	}

	regions, err := core.GetRegions()
	if err != nil {
		log.Println("Error while getting the regions:")
		log.Println(err)
		handleIndexPageLoadingFromConfig(c)
		return
	}

	for _, service := range services {
		for _, region := range regions {
			if region.ServiceID == service.ID {
				service.Regions = append(service.Regions, *region)
			}
		}
	}

	scheduledMainenance, err := core.GetScheduledMaintenanceByID(id)
	if err != nil {
		internalErrorHandler(c, err)
		return
	}

	if scheduledMainenance == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	c.HTML(http.StatusOK, "scheduledMaintenanceDetail.tmpl", gin.H{
		"owner":                config.Config.Website.Title,
		"backgroundColor":      config.Config.Website.HeaderBgColor,
		"cacheBreaker":         config.Config.Website.CacheBreaker,
		"logo":                 "static/img/logo.svg",
		"mostCriticalStatus":   core.MostCriticalServiceStatus(services, regions),
		"services":             services,
		"scheduledMaintenance": scheduledMainenance,
	})
}

// IncidentHistoryHandler is the html controller for sending the html dashboard
func IncidentHistoryHandler(c *gin.Context) {
	pagination := getPaginationFromQuery(c)

	services, err := core.GetServicesEnabled()
	if err != nil {
		log.Println("Error while getting the services:")
		log.Println(err)
		handleIndexPageLoadingFromConfig(c)
		return
	}

	incidents, err := core.GetIncidents(false, pagination)
	if err != nil {
		log.Println("Error while getting the incidents:")
		log.Println(err)
		handleIndexPageLoadingFromConfig(c)
		return
	}

	regions, err := core.GetRegions()
	if err != nil {
		log.Println("Error while getting the regions:")
		log.Println(err)
		handleIndexPageLoadingFromConfig(c)
		return
	}

	for _, service := range services {
		for _, region := range regions {
			if region.ServiceID == service.ID {
				service.Regions = append(service.Regions, *region)
			}
		}
	}

	c.HTML(http.StatusOK, "incidentHistory.tmpl", gin.H{
		"owner":              config.Config.Website.Title,
		"backgroundColor":    config.Config.Website.HeaderBgColor,
		"cacheBreaker":       config.Config.Website.CacheBreaker,
		"logo":               "static/img/logo.svg",
		"services":           services,
		"mostCriticalStatus": core.MostCriticalServiceStatus(services, regions),
		"incidents":          core.AggregateIncidents(incidents, false),
		"page":               pagination.Page,
		"previousPage":       pagination.Page - 1,
		"nextPage":           pagination.Page + 1,
	})
}

func getPaginationFromQuery(c *gin.Context) models.Pagination {
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")
	pageStr := c.Query("page")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 25
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 0
	}

	if page > 0 {
		offset = page * limit
	}

	return models.Pagination{
		Limit:  limit,
		Offset: offset,
		Page:   page,
	}
}
