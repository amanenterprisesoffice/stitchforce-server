package v1

import (
	"errors"
	"log"
	"net/http"
	"stitchforce-server/internal/repo"

	"github.com/gin-gonic/gin"
)

func writeError(c *gin.Context, err error) {
	log.Println("Error On Server: ", err)
	switch {
	case errors.Is(err, repo.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
	}
}
func writeValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

