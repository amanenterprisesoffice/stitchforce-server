package v1

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DateLayout = "2006-01-02"
	CookieMaxAge = 30 * 24 * 60 * 60 // 30 days in seconds
)

func parseDate(c *gin.Context, key string) (*time.Time, error) {
	return parseDateValue(c.Query(key), key)
}

func parseDateValue(value, key string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(DateLayout, value)
	if err != nil {
		return nil, fmt.Errorf("%s must be in YYYY-MM-DD format", key)
	}

	return &parsed, nil
}