package input

type UpdateOrganizationInput struct {
	Name         *string
	AddressLine1 *string
	AddressLine2 *string
	City         *string
	State        *string
	PostalCode   *string
	Country      *string
	Photo        *UploadFile
}