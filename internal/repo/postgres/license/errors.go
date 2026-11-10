package license

import "errors"

var (
	ErrLicenseNotFound      = errors.New("license not found")
	ErrLicenseAlreadyUsed   = errors.New("license already activated")
	ErrLicenseRevoked       = errors.New("license has been revoked")
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrOrganizationSuspended = errors.New("organization is suspended")
	ErrLicenseOrganizationMismatch = errors.New("license does not belong to this organization")
)