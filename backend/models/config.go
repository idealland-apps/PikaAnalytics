package models

import (
    "time"
)

type SystemConfig struct {
    ConfigKey   string    `json:"config_key" db:"config_key"`
    ConfigValue string    `json:"config_value" db:"config_value"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateConfigRequest struct {
    ConfigValue string `json:"config_value" binding:"required"`
}
