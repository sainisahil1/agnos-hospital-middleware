package controllers

import (
	"agnos-hospital-middleware/config"
	middleware "agnos-hospital-middleware/middlewares"
	"agnos-hospital-middleware/models"
	"agnos-hospital-middleware/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testRouter *gin.Engine

func TestMain(m *testing.M) {
	// Set up in-memory test DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to test DB: %v", err)
	}
	config.DB = db
	// Migrate schemas
	err = db.AutoMigrate(&models.Staff{}, &models.Hospital{}, &models.Patient{})
	if err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}
	// Set up router
	gin.SetMode(gin.TestMode)
	testRouter = gin.Default()
	testRouter.POST("/staff/create", CreateStaff)
	testRouter.POST("/staff/login", LoginStaff)
	testRouter.POST("/patient/search", middleware.AuthMiddleware(), SearchPatient)

	code := m.Run()
	os.Exit(code)
}

func TestCreateStaff(t *testing.T) {
	input := models.StaffInput{
		Username: "testuser",
		Password: "secret123",
		Hospital: "Test Hospital",
	}
	body, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/staff/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestCreateStaff_InvalidInput(t *testing.T) {
    invalidInput := "not a valid staff input"
    body, _ := json.Marshal(invalidInput)
    req, _ := http.NewRequest("POST", "/staff/create", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    testRouter.ServeHTTP(w, req)
    assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginStaff(t *testing.T) {
	// Reuse same input to login
	input := models.StaffInput{
		Username: "testuser",
		Password: "secret123",
		Hospital: "Test Hospital",
	}
	body, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/staff/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if _, exists := response["token"]; !exists {
		t.Error("Expected token in response")
	}
}

func TestLoginStaff_InvalidCredentials(t *testing.T) {
    // Test with wrong password
    input := models.StaffInput{
        Username: "testuser",
        Password: "wrongpassword",
        Hospital: "Test Hospital",
    }
    body, _ := json.Marshal(input)
    req, _ := http.NewRequest("POST", "/staff/login", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    testRouter.ServeHTTP(w, req)
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLoginStaff_NonexistentHospital(t *testing.T) {
    input := models.StaffInput{
        Username: "testuser",
        Password: "secret123",
        Hospital: "Nonexistent Hospital",
    }

    body, _ := json.Marshal(input)
    req, _ := http.NewRequest("POST", "/staff/login", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    testRouter.ServeHTTP(w, req)
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}


func TestSearchPatient_ValidSearch(t *testing.T) {
	patient := models.Patient{
		FirstNameEN: "John",
		LastNameEN:  "Doe",
		NationalID:  "12345",
		HospitalID:  1,
	}
	config.DB.Create(&patient)
	hospital := models.Hospital{
		ID: 1,
		Name: "Test Hospital",
	}
	token, err := utils.GenerateJWT(patient.FirstNameEN, hospital)
	if err != nil {
		t.Error("Token generation failed")
		return
	}
	w := httptest.NewRecorder()
	input := PatientSearchInput{
		NationalID: patient.NationalID,
	}
	body, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/patient/search", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	testRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check that the response contains the patient data
	patients := response["patients"].([]any)
	assert.Equal(t, "John", patients[0].(map[string]any)["FirstNameEN"])
	assert.Equal(t, "Doe", patients[0].(map[string]any)["LastNameEN"])

}

// Test case for searching a patient when no local data is found, but external API fetches data
func TestSearchPatient_ExternalFetch(t *testing.T) {
	patient := models.Patient{
		FirstNameEN: "John",
		LastNameEN:  "Doe",
		NationalID:  "12345",
		HospitalID:  1,
	}
	config.DB.Create(&patient)
	hospital := models.Hospital{
		ID: 1,
		Name: "Test Hospital",
	}
	token, err := utils.GenerateJWT(patient.FirstNameEN, hospital)
	if err != nil {
		t.Error("Token generation failed")
		return
	}
	w := httptest.NewRecorder()
	input := PatientSearchInput{
		NationalID: "98765",
	}
	body, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/patient/search", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	testRouter.ServeHTTP(w, req)
	// Assert not equal since it is dependent on external API
	assert.NotEqual(t, http.StatusOK, w.Code)
}

// Test case for missing required input (both NationalID and PassportID missing)
func TestSearchPatient_MissingRequiredFields(t *testing.T) {
	w := httptest.NewRecorder()
	hospital := models.Hospital{
		ID: 1,
		Name: "Test Hospital",
	}
	token, err := utils.GenerateJWT("John", hospital)
	if err != nil {
		t.Error("Token generation failed")
		return
	}
	// Prepare invalid input (no NationalID or PassportID)
	input := PatientSearchInput{}
	body, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/patient/search", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Call the handler
	testRouter.ServeHTTP(w, req)
	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	// Check that the error message is correct
	assert.Equal(t, "At least national_id or passport_id must be provided", response["error"])
}

func TestSearchPatient_InvalidJSON(t *testing.T) {
	hospital := models.Hospital{
		ID: 1,
		Name: "Test Hospital",
	}

	token, err := utils.GenerateJWT("John", hospital)
	if err != nil {
		t.Error("Token generation failed")
		return
	}
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/patient/search", bytes.NewBuffer([]byte("invalid json")))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    testRouter.ServeHTTP(w, req)
    assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchPatient_MissingTokenClaims(t *testing.T) {
    // Create a request without setting the hospital in context
    w := httptest.NewRecorder()
    input := PatientSearchInput{NationalID: "12345"}
    body, _ := json.Marshal(input)
    req, _ := http.NewRequest("POST", "/patient/search", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    // Create a new router without auth middleware for this test
    tempRouter := gin.Default()
    tempRouter.POST("/patient/search", SearchPatient)
    tempRouter.ServeHTTP(w, req)

    assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSearchPatient_PassportIDSearch(t *testing.T) {
    // Create a patient with passport ID
    patient := models.Patient{
        FirstNameEN: "Jane",
        LastNameEN:  "Smith",
        PassportID:  "P98765",
        HospitalID:  1,
    }
    config.DB.Create(&patient)

    hospital := models.Hospital{ID: 1, Name: "Test Hospital"}
    token, _ := utils.GenerateJWT("testuser", hospital)

    w := httptest.NewRecorder()
    input := PatientSearchInput{PassportID: "P98765"}
    body, _ := json.Marshal(input)
    req, _ := http.NewRequest("POST", "/patient/search", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    testRouter.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    patients := response["patients"].([]interface{})
    assert.Equal(t, "Jane", patients[0].(map[string]interface{})["FirstNameEN"])
}

type MockHTTPTransport struct {
    RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    return m.RoundTripFunc(req)
}

func TestSearchPatient_ExternalAPISuccess(t *testing.T) {
    // Setup
    hospital := models.Hospital{ID: 1, Name: "Test Hospital"}
    token, _ := utils.GenerateJWT("testuser", hospital)

    // Create a custom HTTP client with our mock transport
    mockTransport := &MockHTTPTransport{
        RoundTripFunc: func(req *http.Request) (*http.Response, error) {
            // Verify the request URL
            assert.Contains(t, req.URL.String(), "hospital-a.api.co.th/patient/search/98765")

            // Create mock response
            externalPatient := models.PatientExternal{
                FirstNameEN: "External",
                LastNameEN:  "Patient",
                NationalID: "98765",
                PatientHN:  "Test Hospital",
                // Add other required fields
                FirstNameTH:  "ชื่อ",
                LastNameTH:   "นามสกุล",
                DateOfBirth:  "2000-01-01",
                PhoneNumber:  "1234567890",
                Email:        "test@example.com",
                Gender:       "M",
            }
            body, _ := json.Marshal(externalPatient)
            return &http.Response{
                StatusCode: http.StatusOK,
                Body:       io.NopCloser(bytes.NewReader(body)),
                Header:     make(http.Header),
            }, nil
        },
    }

    // Replace the default transport with our mock
    oldTransport := http.DefaultTransport
    http.DefaultTransport = mockTransport
    defer func() { http.DefaultTransport = oldTransport }()

    // Test
    w := httptest.NewRecorder()
    input := PatientSearchInput{NationalID: "98765"}
    body, _ := json.Marshal(input)
    req, _ := http.NewRequest("POST", "/patient/search", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    testRouter.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)

    // Verify response
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    patients := response["patients"].([]interface{})
    assert.Greater(t, len(patients), 0)

    // Verify patient was stored in DB
    var patient models.Patient
    result := config.DB.Where("national_id = ?", "98765").First(&patient)
    assert.Nil(t, result.Error)
    assert.Equal(t, "External", patient.FirstNameEN)
}

func TestSearchPatient_ExternalAPIError(t *testing.T) {
    // Setup
    hospital := models.Hospital{ID: 1, Name: "Test Hospital"}
    token, _ := utils.GenerateJWT("testuser", hospital)

    // Create a custom HTTP client with our mock transport
    mockTransport := &MockHTTPTransport{
        RoundTripFunc: func(req *http.Request) (*http.Response, error) {
            return nil, fmt.Errorf("mock HTTP error")
        },
    }

    // Replace the default transport with our mock
    oldTransport := http.DefaultTransport
    http.DefaultTransport = mockTransport
    defer func() { http.DefaultTransport = oldTransport }()

    // Test
    w := httptest.NewRecorder()
    input := PatientSearchInput{NationalID: "90909"}
    body, _ := json.Marshal(input)
    req, _ := http.NewRequest("POST", "/patient/search", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    testRouter.ServeHTTP(w, req)
    assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSearchPatient_ExternalAPIWrongHospital(t *testing.T) {
    // Setup
	config.DB.Where("1 = 1").Delete(&models.Patient{})

    hospital := models.Hospital{ID: 1, Name: "Wrong Test Hospital"}
    token, _ := utils.GenerateJWT("testuser", hospital)

    // Create a custom HTTP client with our mock transport
    mockTransport := &MockHTTPTransport{
        RoundTripFunc: func(req *http.Request) (*http.Response, error) {
            externalPatient := models.PatientExternal{
                FirstNameEN: "External",
                LastNameEN:  "Patient",
                NationalID: "90987",
                PatientHN:  "Different Hospital",
                FirstNameTH:  "ชื่อ",
                LastNameTH:   "นามสกุล",
                DateOfBirth:  "2000-01-01",
                PhoneNumber:  "1234567890",
                Email:        "test@example.com",
                Gender:       "M",
            }
            body, _ := json.Marshal(externalPatient)
            return &http.Response{
                StatusCode: http.StatusOK,
                Body:       io.NopCloser(bytes.NewReader(body)),
                Header:     make(http.Header),
            }, nil
        },
    }

    // Replace the default transport with our mock
    oldTransport := http.DefaultTransport
    http.DefaultTransport = mockTransport
    defer func() { http.DefaultTransport = oldTransport }()

    // Test
    w := httptest.NewRecorder()
    input := PatientSearchInput{NationalID: "98765"}
    body, _ := json.Marshal(input)
    req, _ := http.NewRequest("POST", "/patient/search", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    testRouter.ServeHTTP(w, req)
    assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStoreInDB_Failure(t *testing.T) {
    // Create a patient that would violate constraints
    invalidPatient := models.Patient{} // Missing required fields
    
    oldDB := config.DB
    defer func() { config.DB = oldDB }()
    
    // Use a new in-memory DB
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    config.DB = db
    
    // Should fail due to missing required fields
    stored := storeInDB(invalidPatient)
    assert.False(t, stored)
}

// Test LoginStaff staff not found
func TestLoginStaff_StaffNotFound(t *testing.T) {
    // Setup - ensure no staff exists
    config.DB.Where("1 = 1").Delete(&models.Staff{})
    
    w := httptest.NewRecorder()
    input := models.StaffInput{
        Username: "nonexistent",
        Password: "password",
        Hospital: "Test Hospital",
    }
    body, _ := json.Marshal(input)
    req, _ := http.NewRequest("POST", "/staff/login", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    testRouter.ServeHTTP(w, req)
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test getClaimsFromToken type assertion error
func TestGetClaimsFromToken_InvalidType(t *testing.T) {
    // Setup gin context with wrong type
    c, _ := gin.CreateTestContext(httptest.NewRecorder())
    c.Set("hospital", "invalid-type") // Not *utils.Claims
    
    _, err := getClaimsFromToken(c)
    assert.Equal(t, http.StatusUnauthorized, err.Code)
}

// Test callExternalAPI response parse error
func TestCallExternalAPI_ParseError(t *testing.T) {
    claims := &utils.Claims{HospitalName: "Test Hospital"}
    
    // Setup mock transport with invalid JSON
    oldTransport := http.DefaultTransport
    defer func() { http.DefaultTransport = oldTransport }()
    
    http.DefaultTransport = &MockHTTPTransport{
        RoundTripFunc: func(req *http.Request) (*http.Response, error) {
            return &http.Response{
                StatusCode: http.StatusOK,
                Body:       io.NopCloser(bytes.NewReader([]byte("invalid json"))),
            }, nil
        },
    }
    
    _, err := callExternalAPI(PatientSearchInput{NationalID: "123"}, claims)
    assert.Equal(t, http.StatusInternalServerError, err.Code)
}