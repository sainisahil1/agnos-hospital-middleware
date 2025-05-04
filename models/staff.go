package models

type Staff struct {
	ID         uint   `gorm:"primaryKey"`
	Username   string `gorm:"unique;not null"`
	Password   string `gorm:"not null"`
	HospitalID uint
	Hospital   Hospital `gorm:"foreignKey:HospitalID"`
}