package handlers

import (
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
)

// GetVersion retrieves the version information from version file
func GetVersion(c *gin.Context) {
	// Read version from version.txt file if it exists
	version := "unknown"
	data, err := os.ReadFile("version.txt")
	if err == nil {
		version = string(data)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"version": version,
	})
}
