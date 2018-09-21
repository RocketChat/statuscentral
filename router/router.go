package router

import (
	"github.com/RocketChat/statuspage/config"
	"github.com/RocketChat/statuspage/controllers"
	"github.com/RocketChat/statuspage/router/middleware"
	"github.com/gin-gonic/gin"
)

//ShowConfig is just a temporary route
func ShowConfig(c *gin.Context) {
	c.JSON(200, config.Config)
}

//Start configures the routes and their handlers plus starts routing
func Start() error {
	router := gin.Default()

	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	router.GET("/", controllers.IndexHandler)

	router.GET("/config", ShowConfig)

	v1 := router.Group("/api").Group("/v1")

	v1.GET("/services", middleware.NotImplemented)
	v1.GET("/incidents", middleware.NotImplemented)
	v1.GET("/incidents/:id/updates", middleware.NotImplemented)

	v1.Use(middleware.IsAuthorized)
	{
		v1.POST("/services", middleware.NotImplemented)
		v1.GET("/services/:id", middleware.NotImplemented)
		v1.PATCH("/services/:id", middleware.NotImplemented)
		v1.DELETE("/services/:id", middleware.NotImplemented)

		v1.POST("/incidents", middleware.NotImplemented)
		v1.GET("/incidents/:id", middleware.NotImplemented)
		v1.DELETE("/incidents/:id", middleware.NotImplemented)

		v1.POST("/incidents/:id/updates", middleware.NotImplemented)
		v1.GET("/incidents/:id/updates/:updateId", middleware.NotImplemented)
		v1.DELETE("/incidents/:id/updates/:updateId", middleware.NotImplemented)
	}

	return router.Run(":5000")
}
