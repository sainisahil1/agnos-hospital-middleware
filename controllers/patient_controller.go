package controllers

import (
	"agnos-hospital-middleware/config"
	"agnos-hospital-middleware/models"
	"agnos-hospital-middleware/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"github.com/gin-gonic/gin"
)

type PatientSearchInput struct {
	NationalID  string `json:"national_id"`
	PassportID  string `json:"passport_id"`
	FirstName   string `json:"first_name"`
	MiddleName  string `json:"middle_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}

type CodedError struct {
	Code  int
	Error string
}

/*
First search for the patient in local db, and return in a slice
If not available, fetch from external API and save in DB, then append to the slice and return
*/
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

	claims, fetchErr := getClaimsFromToken(c)
	if fetchErr != (CodedError{}) {
		c.JSON(fetchErr.Code, gin.H{"error": fetchErr.Error})
		return
	}

	var patients []models.Patient
	patients = fetchFromLocal(claims, input)

	if len(patients) == 0 {
		internalPatient, err := callExternalAPI(input, claims)
		if err != (CodedError{}) {
			c.JSON(err.Code, gin.H{"error": err.Error})
			return
		} else if internalPatient != (models.Patient{}) {
			patients = append(patients, internalPatient)
		}
		// Store in DB
		stored := storeInDB(internalPatient)
		if !stored {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving patient to database"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Success",
		"patients": patients,
	})
}

func storeInDB(internalPatient models.Patient) bool {
	if err := config.DB.Create(&internalPatient).Error; err != nil {
		return false
	}
	return true
}

func callExternalAPI(input PatientSearchInput, claims *utils.Claims) (models.Patient, CodedError) {
	var apiURL string
	if input.NationalID != "" {
		apiURL = fmt.Sprintf("https://hospital-a.api.co.th/patient/search/%s", input.NationalID)
	} else if input.PassportID != "" {
		apiURL = fmt.Sprintf("https://hospital-a.api.co.th/patient/search/%s", input.PassportID)
	}

	resp, err := http.Get(apiURL)
	if err != nil {
		return models.Patient{}, CodedError{Code: http.StatusInternalServerError, Error: "Failed to fetch patient data from external API"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.Patient{}, CodedError{Code: http.StatusInternalServerError, Error: "Error fetching data from external API"}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Patient{}, CodedError{Code: http.StatusInternalServerError, Error: "Failed to read response body"}
	}

	var externalPatient models.PatientExternal
	if err := json.Unmarshal(body, &externalPatient); err != nil {
		return models.Patient{}, CodedError{Code: http.StatusInternalServerError, Error: "Failed to parse external API response"}
	}

	if externalPatient.PatientHN != claims.HospitalName {
		return models.Patient{}, CodedError{Code: http.StatusForbidden, Error: "Access denied for this hospital"}
	}

	Hospital := models.Hospital{
		ID: claims.HospitalID,
		Name: claims.HospitalName,
	}

	internalPatient := models.Patient{
		FirstNameTH:  externalPatient.FirstNameTH,
		MiddleNameTH: externalPatient.MiddleNameTH,
		LastNameTH:   externalPatient.LastNameTH,
		FirstNameEN:  externalPatient.FirstNameEN,
		MiddleNameEN: externalPatient.MiddleNameEN,
		LastNameEN:   externalPatient.LastNameEN,
		DateOfBirth:  externalPatient.DateOfBirth,
		NationalID:   externalPatient.NationalID,
		PassportID:   externalPatient.PassportID,
		PhoneNumber:  externalPatient.PhoneNumber,
		Email:        externalPatient.Email,
		Gender:       externalPatient.Gender,
		HospitalID:   Hospital.ID,
		Hospital:     Hospital,
	}
	return internalPatient, CodedError{}
}

func fetchFromLocal(claims *utils.Claims, input PatientSearchInput) []models.Patient {

	var patients []models.Patient
	query := config.DB.Where("hospital_id = ?", claims.HospitalID) //restricting access to hospital same as staff

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

	result := query.Find(&patients)
	if result.Error != nil {
		//Do nothing. just return empty slice
	}

	return patients
}

func getClaimsFromToken(c *gin.Context) (*utils.Claims, CodedError) {
	hospital, exists := c.Get("hospital")
	if !exists {
		return nil, CodedError{Code: http.StatusUnauthorized, Error: "Hospital info not found in token"}
	}
	//validating the type for claims
	claims, ok := hospital.(*utils.Claims)
	if !ok {
		return nil, CodedError{Code: http.StatusUnauthorized, Error: "Invalid hospital claims type"}
	}
	return claims, CodedError{}
}
