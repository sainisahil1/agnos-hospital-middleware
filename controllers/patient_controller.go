package controllers

import (
	"agnos-hospital-middleware/config"
	"agnos-hospital-middleware/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PatientSearchInput struct {
	NationalID   string `json:"national_id"`
	PassportID   string `json:"passport_id"`
	FirstName    string `json:"first_name"`
	MiddleName   string `json:"middle_name"`
	LastName     string `json:"last_name"`
	DateOfBirth  string `json:"date_of_birth"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
}

func SearchPatient(c *gin.Context) {
	//Get input from the request
	var input PatientSearchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that at least one field is provided for searching
	if input.NationalID == "" && input.PassportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least national_id or passport_id must be provided"})
		return
	}

	hospital, exists := c.Get("hospital")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Hospital info not found in token"})
		return
	}

	// Form the URL to hit the external hospital API
	var apiURL string
	if input.NationalID != "" {
		apiURL = fmt.Sprintf("https://hospital-a.api.co.th/patient/search/%s", input.NationalID)
	} else if input.PassportID != "" {
		apiURL = fmt.Sprintf("https://hospital-a.api.co.th/patient/search/%s", input.PassportID)
	}

	// Call the external API
	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch patient data from external API"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Error fetching data from external API"})
		return
	}








	var patients []models.Patient
	query := config.DB.Where("hospital = ?", hospital)

	if input.NationalID != "" {
		query = query.Where("national_id = ?", input.NationalID)
	}
	if input.PassportID != "" {
		query = query.Where("passport_id = ?", input.PassportID)
	}
	if input.FirstName != "" {
		query = query.Where("first_name_en = ?", input.FirstName)
	}
	if input.MiddleName != "" {
		query = query.Where("middle_name_en = ?", input.MiddleName)
	}
	if input.LastName != "" {
		query = query.Where("last_name_en = ?", input.LastName)
	}
	if input.DateOfBirth != "" {
		query = query.Where("date_of_birth = ?", input.DateOfBirth)
	}
	if input.PhoneNumber != "" {
		query = query.Where("phone_number = ?", input.PhoneNumber)
	}
	if input.Email != "" {
		query = query.Where("email = ?", input.Email)
	}

	if err := query.Find(&patients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed"})
		return
	}

	c.JSON(http.StatusOK, patients)
}
