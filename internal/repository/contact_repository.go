package repository

import (
	"context"
	"errors"

	"contactmanagement/internal/models"
	"contactmanagement/internal/types"

	"gorm.io/gorm"
)

// ContactRepository defines the interface for contact persistence operations
type ContactRepository interface {
	Create(ctx context.Context, contact *models.Contact) error
	FindByID(ctx context.Context, id uint) (*models.Contact, error)
	Update(ctx context.Context, contact *models.Contact) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]models.Contact, error)
	FindByEmail(ctx context.Context, email string) (*models.Contact, error)
	ImportContacts(ctx context.Context, contacts []models.Contact) error
	CheckDuplicatePhone(ctx context.Context, phoneNumber string, excludeContactID uint) (bool, error)
}

// GormContactRepository implements ContactRepository using GORM
type GormContactRepository struct {
	db *gorm.DB
}

// NewContactRepository creates a new contact repository
func NewContactRepository(db *gorm.DB) ContactRepository {
	return &GormContactRepository{db: db}
}

// CheckDuplicatePhone checks if a phone number exists in any contact
func (r *GormContactRepository) CheckDuplicatePhone(ctx context.Context, phoneNumber string, excludeContactID uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&types.Contact{}).
		Where("primary_phone = ?", phoneNumber)

	if excludeContactID > 0 {
		query = query.Where("id != ?", excludeContactID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	// Check in additional phones
	query = r.db.WithContext(ctx).Model(&types.Phone{}).
		Where("number = ?", phoneNumber)

	if excludeContactID > 0 {
		query = query.Where("contact_id != ?", excludeContactID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Create stores a new contact and its additional phones
func (r *GormContactRepository) Create(ctx context.Context, contact *models.Contact) error {
	// Check for duplicate primary phone
	exists, err := r.CheckDuplicatePhone(ctx, contact.PrimaryPhone, 0)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("primary phone number already exists in another contact")
	}

	// Check for duplicate additional phones
	for _, phone := range contact.AdditionalPhones {
		exists, err := r.CheckDuplicatePhone(ctx, phone.Number, 0)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("additional phone number already exists in another contact")
		}
	}

	// Save the contact
	if err := r.db.WithContext(ctx).Create(contact.Contact).Error; err != nil {
		return err
	}

	// Save additional phones
	for i := range contact.AdditionalPhones {
		contact.AdditionalPhones[i].ContactID = contact.ID
		contact.AdditionalPhones[i].ID = 0
	}
	if len(contact.AdditionalPhones) > 0 {
		if err := r.db.WithContext(ctx).Create(&contact.AdditionalPhones).Error; err != nil {
			return err
		}
	}

	return nil
}

// FindByID retrieves a contact by ID
func (r *GormContactRepository) FindByID(ctx context.Context, id uint) (*models.Contact, error) {
	var contact models.Contact
	if err := r.db.WithContext(ctx).Preload("AdditionalPhones").First(&contact, id).Error; err != nil {
		return nil, err
	}
	return &contact, nil
}

// Update modifies an existing contact and its additional phones
func (r *GormContactRepository) Update(ctx context.Context, contact *models.Contact) error {
	// Check for duplicate primary phone
	exists, err := r.CheckDuplicatePhone(ctx, contact.PrimaryPhone, contact.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("primary phone number already exists in another contact")
	}

	// Check for duplicate additional phones
	for _, phone := range contact.AdditionalPhones {
		exists, err := r.CheckDuplicatePhone(ctx, phone.Number, contact.ID)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("additional phone number already exists in another contact")
		}
	}

	// Update the contact
	if err := r.db.WithContext(ctx).Save(contact.Contact).Error; err != nil {
		return err
	}

	// Delete old additional phones
	if err := r.db.WithContext(ctx).Unscoped().Where("contact_id = ?", contact.ID).Delete(&types.Phone{}).Error; err != nil {
		return err
	}

	// Save new additional phones
	for i := range contact.AdditionalPhones {
		contact.AdditionalPhones[i].ContactID = contact.ID
		contact.AdditionalPhones[i].ID = 0
	}
	if len(contact.AdditionalPhones) > 0 {
		if err := r.db.WithContext(ctx).Create(&contact.AdditionalPhones).Error; err != nil {
			return err
		}
	}

	return nil
}

// Delete removes a contact and its additional phones permanently from the database
func (r *GormContactRepository) Delete(ctx context.Context, id uint) error {
	// First delete all associated phone numbers (hard delete)
	if err := r.db.WithContext(ctx).Unscoped().Where("contact_id = ?", id).Delete(&types.Phone{}).Error; err != nil {
		return err
	}
	// Then delete the contact (hard delete)
	return r.db.WithContext(ctx).Unscoped().Delete(&types.Contact{}, id).Error
}

// List retrieves all contacts
func (r *GormContactRepository) List(ctx context.Context) ([]models.Contact, error) {
	var contacts []models.Contact
	if err := r.db.WithContext(ctx).Preload("AdditionalPhones").Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}

// FindByEmail finds a contact by email address
func (r *GormContactRepository) FindByEmail(ctx context.Context, email string) (*models.Contact, error) {
	var contact models.Contact
	if err := r.db.WithContext(ctx).Preload("AdditionalPhones").Where("email = ?", email).First(&contact).Error; err != nil {
		return nil, err
	}
	return &contact, nil
}

// ImportContacts imports multiple contacts
func (r *GormContactRepository) ImportContacts(ctx context.Context, contacts []models.Contact) error {
	return r.db.WithContext(ctx).Create(contacts).Error
}
