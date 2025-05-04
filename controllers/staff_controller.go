package controllers

import (
	"agnos-hospital-middleware/config"
	"agnos-hospital-middleware/models"
	"agnos-hospital-middleware/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
-> Find the hospital with input, create if does not exist
-> hash the password
-> create the staff
-> push to DB
*/
func CreateStaff(c *gin.Context) {
	var input models.StaffInput
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
		Username:   input.Username,
		Password:   hashedPassword,
		HospitalID: hospital.ID,
	}

	if err := db.Create(&staff).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create staff"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Staff created"})

}

/*
-> fetch staff from DB
-> match the password hash with input
-> Generate JWT
*/
func LoginStaff(c *gin.Context) {
	var input models.StaffInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := config.DB

	var hospital models.Hospital
	if err := db.Where("name = ?", input.Hospital).First(&hospital).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	var staff models.Staff
	if err := db.Where("username = ? AND hospital_id = ?", input.Username, hospital.ID).First(&staff).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !utils.CheckPasswordHash(input.Password, staff.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := utils.GenerateJWT(staff.Username, staff.Hospital)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
