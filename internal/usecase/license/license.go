package license

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"stitchforce-server/internal/entity"
	"stitchforce-server/internal/repo"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/pkg/logger"
)

type UseCase struct {
	repository repo.LicenseRepository
	logger     *logger.Logger
}

func New(repository repo.LicenseRepository, logger *logger.Logger) usecase.LicenseUseCase {
	return &UseCase{
		repository: repository,
		logger:     logger,
	}
}

type licensePayload struct {
	OrganizationID string `json:"org"`
	DurationDays   int    `json:"days"`
	Nonce          string `json:"nonce"`
}

func (u *UseCase) Generate(
	ctx context.Context,
	organizationID string,
	durationDays int,
) (*entity.GeneratedLicense, error) {
	if strings.TrimSpace(organizationID) == "" {
		return nil, errors.New("organization ID is required")
	}

	if durationDays <= 0 || durationDays > 3650 {
		return nil, errors.New("duration must be between 1 and 3650 days")
	}

	secret := []byte(os.Getenv("LICENSE_SIGNING_SECRET"))
	if len(secret) < 32 {
		return nil, errors.New("LICENSE_SIGNING_SECRET must be at least 32 bytes")
	}

	// Generate a cryptographically secure random nonce.
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	payload := licensePayload{
		OrganizationID: organizationID,
		DurationDays:   durationDays,
		Nonce:          base64.RawURLEncoding.EncodeToString(nonce),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// Sign the payload using HMAC-SHA256.
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(encodedPayload))

	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	licenseKey := "STF-" + encodedPayload + "-" + signature

	// Store only the SHA-256 hash of the complete license key.
	hash := sha256.Sum256([]byte(licenseKey))
	hashString := hex.EncodeToString(hash[:])

	license := &entity.License{
		LicenseKeyHash: hashString,
		OrganizationID: &organizationID,
		DurationDays:   durationDays,
		Status:         "unused",
	}

	if err := u.repository.Create(ctx, license); err != nil {
		return nil, err
	}

	return &entity.GeneratedLicense{
		License:    license,
		LicenseKey: licenseKey,
	}, nil
}

func (u *UseCase) Activate(
	ctx context.Context,
	organizationID string,
	licenseKey string,
) (*entity.License, error) {
	licenseKey = strings.TrimSpace(licenseKey)

	if licenseKey == "" {
		return nil, errors.New("license key is required")
	}

	secret := []byte(os.Getenv("LICENSE_SIGNING_SECRET"))
	if len(secret) < 32 {
		return nil, errors.New("LICENSE_SIGNING_SECRET must be at least 32 bytes")
	}

	// Verify key format and signature.
	parts := strings.Split(licenseKey, "-")
	if len(parts) != 3 || parts[0] != "STF" {
		return nil, errors.New("invalid license key")
	}

	encodedPayload := parts[1]
	encodedSignature := parts[2]

	signature, err := base64.RawURLEncoding.DecodeString(encodedSignature)
	if err != nil {
		return nil, errors.New("invalid license signature")
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(encodedPayload))

	if !hmac.Equal(signature, mac.Sum(nil)) {
		return nil, errors.New("invalid license signature")
	}

	// Decode the signed payload.
	payloadBytes, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return nil, errors.New("invalid license payload")
	}

	var payload licensePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, errors.New("invalid license payload")
	}

	// Verify the license was generated for this organization.
	if payload.OrganizationID != organizationID {
		return nil, errors.New("license does not belong to this organization")
	}

	if payload.DurationDays <= 0 || payload.Nonce == "" {
		return nil, errors.New("invalid license payload")
	}

	// Hash the exact key. Do not uppercase it: Base64 URL encoding is case-sensitive.
	hash := sha256.Sum256([]byte(licenseKey))
	hashString := hex.EncodeToString(hash[:])

	return u.repository.Activate(
		ctx,
		hashString,
		organizationID,
	)
}
