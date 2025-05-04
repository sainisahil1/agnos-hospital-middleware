package routes

import (
	"agnos-hospital-middleware/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.POST("/staff/create", controllers.CreateStaff)
	r.POST("/staff/login", controllers.LoginStaff)

}
