package models

import (
	"errors"
	"regexp"
	"strings"

	"contactmanagement/internal/types"
)

// Contact represents a contact with business logic
type Contact struct {
	*types.Contact
}

// NewContact creates a new contact model
func NewContact(contact *types.Contact) *Contact {
	return &Contact{Contact: contact}
}

// Validate performs business validation on the contact
func (c *Contact) Validate() error {
	if strings.TrimSpace(c.FirstName) == "" {
		return errors.New("first name is required")
	}
	if strings.TrimSpace(c.LastName) == "" {
		return errors.New("last name is required")
	}
	if !isValidEmail(c.Email) {
		return errors.New("invalid email format")
	}
	if !isValidPhone(c.PrimaryPhone) {
		return errors.New("invalid primary phone format")
	}
	// Prevent duplicate phone numbers (number + type)
	phoneSet := make(map[string]struct{})
	for _, phone := range c.AdditionalPhones {
		key := phone.Number + "-" + phone.Type
		if _, exists := phoneSet[key]; exists {
			return errors.New("duplicate phone numbers are not allowed")
		}
		phoneSet[key] = struct{}{}
	}
	return nil
}

// FormatPhoneNumbers formats all phone numbers in the contact
func (c *Contact) FormatPhoneNumbers() {
	c.PrimaryPhone = formatPhoneNumber(c.PrimaryPhone)
	for i := range c.AdditionalPhones {
		c.AdditionalPhones[i].Number = formatPhoneNumber(c.AdditionalPhones[i].Number)
	}
}

// Helper functions
func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}

func isValidPhone(phone string) bool {
	// Remove all non-numeric characters
	cleaned := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	// Check if the resulting number is 10-15 digits
	return len(cleaned) >= 10 && len(cleaned) <= 15
}

func formatPhoneNumber(phone string) string {
	// Remove all non-numeric characters
	cleaned := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	// Format can be customized based on your needs
	return cleaned
}
