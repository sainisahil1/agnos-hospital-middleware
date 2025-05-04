package models

type Patient struct {
	ID            uint      `gorm:"primaryKey"`
	FirstNameTH   string
	MiddleNameTH  string
	LastNameTH    string
	FirstNameEN   string
	MiddleNameEN  string
	LastNameEN    string
	DateOfBirth   string
	NationalID    string
	PassportID    string
	PhoneNumber   string
	Email         string
	Gender        string
	HospitalID    uint
	Hospital      Hospital `gorm:"foreignKey:HospitalID"`
}
