package types

import "gorm.io/gorm"

// Contact represents a contact in the system
type Contact struct {
	gorm.Model
	FirstName       string `json:"firstName" gorm:"not null"`
	LastName        string `json:"lastName" gorm:"not null"`
	Email           string `json:"email" gorm:"unique;not null"`
	PrimaryPhone    string `json:"primaryPhone" gorm:"not null"`
	AdditionalPhones []Phone `json:"additionalPhones,omitempty" gorm:"foreignKey:ContactID"`
}

// Phone represents a phone number associated with a contact
type Phone struct {
	gorm.Model
	ContactID uint   `json:"contactId"`
	Number    string `json:"number"`
	Type      string `json:"type"` // e.g., "home", "work", "mobile"
}

// CreateContactRequest represents the request body for creating a contact
type CreateContactRequest struct {
	FirstName        string   `json:"firstName" binding:"required"`
	LastName         string   `json:"lastName" binding:"required"`
	Email            string   `json:"email" binding:"required,email"`
	PrimaryPhone     string   `json:"primaryPhone" binding:"required"`
	AdditionalPhones []Phone  `json:"additionalPhones"`
}

// CSVContact represents a contact in CSV format
type CSVContact struct {
	FirstName    string `csv:"First Name"`
	LastName     string `csv:"Last Name"`
	Email        string `csv:"Email Address"`
	PrimaryPhone string `csv:"Primary Phone Number"`
} 