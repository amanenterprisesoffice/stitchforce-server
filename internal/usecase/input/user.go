package input

import "stitchforce-server/internal/entity"

type UpdateUserInput struct {
	Name  *string
	Email *string
	Photo *UploadFile
	Role  *entity.UserRole
}

type UserFilter struct {
	Search string
	Role   string
	Limit  int
	Offset int
}
