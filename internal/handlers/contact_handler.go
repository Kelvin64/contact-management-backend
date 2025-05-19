package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocarina/gocsv"

	"contactmanagement/internal/models"
	"contactmanagement/internal/repository"
	"contactmanagement/internal/types"
)

type ContactHandler struct {
	repo repository.ContactRepository
}

func NewContactHandler(repo repository.ContactRepository) *ContactHandler {
	return &ContactHandler{repo: repo}
}

func (h *ContactHandler) CreateContact(c *gin.Context) {
	var req types.CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a new contact model
	contact := models.NewContact(&types.Contact{
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Email:            req.Email,
		PrimaryPhone:     req.PrimaryPhone,
		AdditionalPhones: req.AdditionalPhones,
	})

	// Validate the contact
	if err := contact.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Format phone numbers
	contact.FormatPhoneNumbers()

	// Check for existing email using repository
	existingContact, err := h.repo.FindByEmail(c.Request.Context(), req.Email)
	if err == nil && existingContact != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// Create contact using repository
	if err := h.repo.Create(c.Request.Context(), contact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contact"})
		return
	}

	c.JSON(http.StatusCreated, contact.Contact)
}

func (h *ContactHandler) ListContacts(c *gin.Context) {
	contacts, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contacts"})
		return
	}

	c.JSON(http.StatusOK, contacts)
}

func (h *ContactHandler) GetContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	contact, err := h.repo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}
	c.JSON(http.StatusOK, contact)
}

func (h *ContactHandler) UpdateContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req types.CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing contact
	existingContact, err := h.repo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}

	// Update contact fields
	existingContact.Contact.FirstName = req.FirstName
	existingContact.Contact.LastName = req.LastName
	existingContact.Contact.Email = req.Email
	existingContact.Contact.PrimaryPhone = req.PrimaryPhone
	existingContact.Contact.AdditionalPhones = req.AdditionalPhones

	// Validate updated contact
	if err := existingContact.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Format phone numbers
	existingContact.FormatPhoneNumbers()

	// Check for duplicate email
	emailContact, err := h.repo.FindByEmail(c.Request.Context(), req.Email)
	if err == nil && emailContact != nil && emailContact.Contact.ID != existingContact.Contact.ID {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// Update using repository
	if err := h.repo.Update(c.Request.Context(), existingContact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contact"})
		return
	}

	c.JSON(http.StatusOK, existingContact)
}

func (h *ContactHandler) DeleteContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Check if contact exists
	_, err = h.repo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}

	// Delete using repository
	if err := h.repo.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete contact"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contact deleted successfully"})
}

func (h *ContactHandler) ImportContacts(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	if !strings.HasSuffix(file.Filename, ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File must be a CSV"})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer openedFile.Close()

	// Read the entire file into a buffer
	fileBytes, err := io.ReadAll(openedFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	reader := bytes.NewReader(fileBytes)
	var csvContacts []types.CSVContact
	if err := gocsv.Unmarshal(reader, &csvContacts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse CSV file"})
		return
	}

	var contactsToImport []models.Contact
	for _, csvContact := range csvContacts {
		contact := models.NewContact(&types.Contact{
			FirstName:    csvContact.FirstName,
			LastName:     csvContact.LastName,
			Email:        csvContact.Email,
			PrimaryPhone: csvContact.PrimaryPhone,
		})

		// Validate each contact
		if err := contact.Validate(); err != nil {
			continue
		}

		// Format phone numbers
		contact.FormatPhoneNumbers()
		contactsToImport = append(contactsToImport, *contact)
	}

	// Import contacts using repository
	if err := h.repo.ImportContacts(c.Request.Context(), contactsToImport); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import contacts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Import completed",
		"imported": len(contactsToImport),
	})
}
