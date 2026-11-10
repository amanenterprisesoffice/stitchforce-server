package request


type ActivateLicenseRequest struct {
	LicenseKey string `json:"license_key" validate:"required"`
}