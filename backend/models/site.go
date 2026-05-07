package models

import "time"

type Site struct {
    ID          int       `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    SiteKey     string    `json:"site_key" db:"site_key"`
    Domain      string    `json:"domain" db:"domain"`
    Description string    `json:"description" db:"description"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateSiteRequest struct {
    Name        string `json:"name" binding:"required"`
    SiteKey     string `json:"site_key" binding:"required"`
    Domain      string `json:"domain"`
    Description string `json:"description"`
}

type UpdateSiteRequest struct {
    Name        string `json:"name"`
    Domain      string `json:"domain"`
    Description string `json:"description"`
}
