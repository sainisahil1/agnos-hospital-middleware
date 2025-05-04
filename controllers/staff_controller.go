package controllers

import (
	"agnos-hospital-middleware/config"
	"agnos-hospital-middleware/models"
	"agnos-hospital-middleware/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StaffInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Hospital string `json:"hospital" binding:"required"`
}

func CreateStaff(c *gin.Context){
	var input StaffInput
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := config.DB

	//Create or find Hospital
	var hospital models.Hospital
	db.FirstOrCreate(&hospital, models.Hospital{Name: input.Hospital})

	//Hash Password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password encryption failed"})
		return
	}

	//Create Staff
	staff := models.Staff{
		Username: input.Username,
		Password: hashedPassword,
		HospitalID: hospital.ID,
	}

	if err := db.Create(&staff).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create staff"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Staff created"})

}


func LoginStaff(c *gin.Context) {
	var input StaffInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := config.DB

	var staff models.Staff
	if err := db.Where("username = ? AND hospital = ?", input.Username, input.Hospital).First(&staff).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if utils.CheckPasswordHash(input.Password, staff.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := utils.GenerateJWT(staff.Username, staff.Hospital.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
