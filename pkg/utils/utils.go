package utils

import (
	"fmt"
	"math"
	"strconv"

	"github.com/google/uuid"
)

func NewID() (string, error) {
	return uuid.NewString(), nil
}

// RupeesToPaise converts a Rupee amount (string to avoid float issues) to Paise int64
func RupeesToPaise(rupeeStr string) (int64, error) {
	// Parse the string into a float just for the initial conversion,
	// then immediately round to the nearest integer to avoid precision loss.
	rupees, err := strconv.ParseFloat(rupeeStr, 64)
	if err != nil {
		return 0, err
	}

	// 1 Rupee = 100 Paise
	paise := math.Round(rupees * 100)
	return int64(paise), nil
}

// PaiseToRupees converts Paise int64 into a formatted Rupee string
func PaiseToRupees(paise int64) string {
	rupees := paise / 100
	remainingPaise := paise % 100

	// Handles negative values correctly for display
	if remainingPaise < 0 {
		remainingPaise = -remainingPaise
	}

	return fmt.Sprintf("%d.%02d", rupees, remainingPaise)
}
