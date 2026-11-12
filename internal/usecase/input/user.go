package input

type UpdateUserInput struct {
	Name  *string
	Email *string
	Photo *UploadFile
}

type UserFilter struct {
	Search string
	Role   string
	Limit  int
	Offset int
}
