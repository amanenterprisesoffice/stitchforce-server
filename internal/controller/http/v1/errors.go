package v1

import (
	"errors"
	"net/http"
	"stitchforce-server/internal/repo"

	"github.com/gin-gonic/gin"
)

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repo.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "not found"}})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
	}
}
func writeValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
