package utils

import (
	"github.com/google/uuid"
)

func NewID() (string, error) {
	return uuid.NewString(), nil
}
