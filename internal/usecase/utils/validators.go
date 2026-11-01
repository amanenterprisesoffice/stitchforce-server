package utils

import (
	"errors"
	"stitchforce-server/internal/entity"
	"strings"
	"unicode/utf8"
)

func ValidateSizeID(id string) (string, error) {
	id = strings.TrimSpace(id)

	if id == "" {
		return "", errors.New("invalid size ID")
	}

	return id, nil
}

func ValidateOrganization(org *entity.Organization) error {
	if org == nil {
		return errors.New("organization is required")
	}

	org.Name = strings.TrimSpace(org.Name)
	length := utf8.RuneCountInString(org.Name)

	if length < 2 || length > 150 {
		return errors.New("organization name must be between 2 and 150 characters")
	}

	return nil
}


func ValidateSize(size *entity.Size) error {
    if size == nil {
        return errors.New("size is nil")
    }

    size.Name = strings.TrimSpace(size.Name)

    if size.Name == "" {
        return errors.New("size name is required")
    }

    if len(size.Name) > 50 {
        return errors.New("size name must not exceed 50 characters")
    }

    return nil
}

func ValidateOrganizationID(organizationID string) (string, error) {
	organizationID = strings.TrimSpace(organizationID)

	if organizationID == "" {
		return "", errors.New("organization id is required")
	}

	return organizationID, nil
}


func ValidateWorker(worker *entity.GarmentWorker) error {
	if worker == nil {
		return errors.New("garment worker is required")
	}

	worker.Name = strings.TrimSpace(worker.Name)
	worker.Phone = strings.TrimSpace(worker.Phone)
	worker.Gender = strings.ToLower(strings.TrimSpace(worker.Gender))

	if worker.Name == "" {
		return errors.New("name is required")
	}

	if worker.Phone == "" {
		return errors.New("phone is required")
	}

	switch worker.Gender {
	case "male", "female", "other":
	default:
		return errors.New("gender must be male, female, or other")
	}

	return nil
}



func ValidateStyle(style *entity.Style) error {
	if style == nil {
		return errors.New("style is required")
	}

	style.Name = strings.TrimSpace(style.Name)
	style.StyleNumber = strings.TrimSpace(style.StyleNumber)
	style.Description = strings.TrimSpace(style.Description)

	if style.Name == "" {
		return errors.New("style name is required")
	}

	if style.StyleNumber == "" {
		return errors.New("style number is required")
	}

	return nil
}

func ValidateStyleID(styleID string) (string, error) {
	styleID = strings.TrimSpace(styleID)

	if styleID == "" {
		return "", errors.New("style id is required")
	}

	return styleID, nil
}


func ValidateStyleOperation(operation *entity.StyleOperation) error {
	if operation == nil {
		return errors.New("style operation is required")
	}

	operation.Name = strings.TrimSpace(operation.Name)

	if operation.Name == "" {
		return errors.New("operation name is required")
	}

	if operation.RatePerPiece < 0 {
		return errors.New("rate per piece cannot be negative")
	}

	return nil
}