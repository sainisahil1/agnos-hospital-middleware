package main

import (
	"agnos-hospital-middleware/config"
	"agnos-hospital-middleware/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize DB
	config.InitDB()

	// Start Gin router
	r := gin.Default()

	// Set up routes
	routes.SetupRoutes(r)

	// Run server
	r.Run("0.0.0.0:8080")
}
