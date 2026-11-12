package request

import "stitchforce-server/internal/entity"

type InviteUser struct {
	Email string          `json:"email" validate:"required,email,max=255"`
	Role  entity.UserRole `json:"role" validate:"required,oneof=org_admin staff"`
}

type AcceptInvitation struct {
	Token    string `json:"token" validate:"required"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}
