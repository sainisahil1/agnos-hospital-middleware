package routes

import (
	"agnos-hospital-middleware/controllers"
	middleware "agnos-hospital-middleware/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.POST("/staff/create", controllers.CreateStaff)
	r.POST("/staff/login", controllers.LoginStaff)

	protected := r.Group("/patient")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/search", controllers.SearchPatient)
	}

}
