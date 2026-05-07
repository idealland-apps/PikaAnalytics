package handlers

import (
    "database/sql"
    "net/http"
    "time"
    "pikaanalytics-backend/database"
    "pikaanalytics-backend/models"
    "github.com/gin-gonic/gin"
)

func GetConfig(c *gin.Context) {
    key := c.Param("key")

    var config models.SystemConfig
    err := database.DB.QueryRow(
        "SELECT config_key, config_value, created_at, updated_at FROM system_config WHERE config_key = ?",
        key,
    ).Scan(&config.ConfigKey, &config.ConfigValue, &config.CreatedAt, &config.UpdatedAt)

    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve configuration"})
        return
    }

    c.JSON(http.StatusOK, config)
}

func UpdateConfig(c *gin.Context) {
    key := c.Param("key")

    var request models.UpdateConfigRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    var exists bool
    if err := database.DB.QueryRow(
        "SELECT EXISTS(SELECT 1 FROM system_config WHERE config_key = ?)", key,
    ).Scan(&exists); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check configuration"})
        return
    }

    var err error
    if exists {
        _, err = database.DB.Exec(
            "UPDATE system_config SET config_value = ?, updated_at = ? WHERE config_key = ?",
            request.ConfigValue, time.Now(), key,
        )
    } else {
        _, err = database.DB.Exec(
            "INSERT INTO system_config (config_key, config_value, created_at, updated_at) VALUES (?, ?, ?, ?)",
            key, request.ConfigValue, time.Now(), time.Now(),
        )
    }

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully"})
}
