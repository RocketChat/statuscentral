package router

import (
	"fmt"

	"github.com/RocketChat/statuscentral/config"
	v1c "github.com/RocketChat/statuscentral/controllers/v1"
	"github.com/RocketChat/statuscentral/router/middleware"
	"github.com/gin-gonic/gin"
)

// Start configures the routes and their handlers plus starts routing
func Start(port int) error {
	runMetricsRouter()

	router := gin.Default()

	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*.tmpl")

	router.GET("/", v1c.IndexHandler)

	router.GET("/incidents/:id", v1c.IncidentDetailHandler)
	router.GET("/i/:id", v1c.IncidentShortRedirectHandler)

	router.GET("/scheduled-maintenance/:id", v1c.ScheduledMaintenanceDetailHandler)
	router.GET("/m/:id", v1c.ScheduledMaintenanceShortRedirectHandler)

	router.GET("/incidents", v1c.IncidentHistoryHandler)

	v1 := router.Group("/api").Group("/v1")

	v1.GET("/services", v1c.ServicesGetAll)
	v1.GET("/incidents", v1c.IncidentsGetAll)
	v1.GET("/incidents/:id/updates", v1c.IncidentUpdatesGetAll)

	v1.GET("/scheduled-maintenance", v1c.ScheduledMaintenanceGetAll)
	v1.GET("/scheduled-maintenance/:id/updates", v1c.ScheduledMaintenanceUpdatesGetAll)

	v1.Use(middleware.IsAuthorized)
	{
		v1.GET("/config", config.Config.HttpHandler)

		// Services
		v1.POST("/services", v1c.ServiceCreate)
		v1.GET("/services/:id", v1c.ServicesGetOne)
		v1.POST("/services/:id", v1c.ServiceUpdate)
		v1.DELETE("/services/:id", middleware.NotImplemented)

		// Regions
		v1.POST("/regions", v1c.RegionCreate)
		v1.GET("/regions/:id", middleware.NotImplemented)
		v1.PATCH("/regions/:id", middleware.NotImplemented)
		v1.DELETE("/regions/:id", v1c.RegionDelete)

		// Incidents
		v1.POST("/incidents", v1c.IncidentCreate)
		v1.GET("/incidents/:id", v1c.IncidentGetOne)
		v1.DELETE("/incidents/:id", v1c.IncidentDelete)

		v1.POST("/incidents/:id/updates", v1c.IncidentUpdateCreate)
		v1.GET("/incidents/:id/updates/:updateId", v1c.IncidentUpdateGetOne)
		v1.DELETE("/incidents/:id/updates/:updateId", v1c.IncidentUpdateDelete)

		// Scheduled Maintenance
		v1.POST("/scheduled-maintenance", v1c.ScheduledMaintenanceCreate)
		v1.GET("/scheduled-maintenance/:id", v1c.ScheduledMaintenanceGetOne)
		v1.PATCH("/scheduled-maintenance/:id", v1c.ScheduledMaintenancePatch)
		v1.DELETE("/scheduled-maintenance/:id", v1c.ScheduledMaintenanceDelete)

		v1.POST("/scheduled-maintenance/:id/updates", v1c.ScheduledMaintenanceUpdateCreate)
		v1.GET("/scheduled-maintenance/:id/updates/:updateId", v1c.ScheduledMaintenanceUpdateGetOne)
		v1.DELETE("/scheduled-maintenance/:id/updates/:updateId", v1c.ScheduledMaintenanceUpdateDelete)
	}

	return router.Run(fmt.Sprintf(":%d", port))
}

func runMetricsRouter() {
	healthMetricsRouter := gin.Default()
	healthMetricsRouter.GET("/health", v1c.LivenessCheckHandler)

	// Endpoint that will return a snapshot of the bolt database. Can be used for backup purposes
	healthMetricsRouter.GET("/snapshot", middleware.IsAuthorized, v1c.SnapshotHandler)

	go healthMetricsRouter.Run(":8080")
}
