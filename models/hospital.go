package models

type Hospital struct{
	ID uint `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}