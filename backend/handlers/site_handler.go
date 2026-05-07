package handlers

import (
    "database/sql"
    "net/http"
    "strings"
    "time"

    "pikaanalytics-backend/database"
    "pikaanalytics-backend/models"
    "github.com/gin-gonic/gin"
)

func isValidSiteKey(key string) bool {
    if len(key) < 1 || len(key) > 64 {
        return false
    }
    for _, r := range key {
        if !(r == '-' || r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
            return false
        }
    }
    return true
}

func isSiteKeyAvailable(siteKey string) bool {
    var count int
    err := database.DB.QueryRow("SELECT COUNT(*) FROM sites WHERE site_key = ?", siteKey).Scan(&count)
    return err == nil && count == 0
}

func GetSites(c *gin.Context) {
    rows, err := database.DB.Query(`
        SELECT id, name, site_key, domain, description, created_at, updated_at
        FROM sites
        ORDER BY created_at DESC
    `)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sites"})
        return
    }
    defer rows.Close()

    var sites []models.Site
    for rows.Next() {
        var site models.Site
        err := rows.Scan(&site.ID, &site.Name, &site.SiteKey, &site.Domain, &site.Description, &site.CreatedAt, &site.UpdatedAt)
        if err != nil {
            continue
        }
        sites = append(sites, site)
    }

    c.JSON(http.StatusOK, sites)
}

func GetSite(c *gin.Context) {
    id := c.Param("id")
    var site models.Site
    err := database.DB.QueryRow(`
        SELECT id, name, site_key, domain, description, created_at, updated_at
        FROM sites
        WHERE id = ?
    `, id).Scan(&site.ID, &site.Name, &site.SiteKey, &site.Domain, &site.Description, &site.CreatedAt, &site.UpdatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "Site not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch site"})
        return
    }

    c.JSON(http.StatusOK, site)
}

func CreateSite(c *gin.Context) {
    var req models.CreateSiteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    req.SiteKey = strings.TrimSpace(req.SiteKey)
    if !isValidSiteKey(req.SiteKey) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "site_key must be 1-64 characters and contain only letters, numbers, hyphens, and underscores"})
        return
    }

    if !isSiteKeyAvailable(req.SiteKey) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "site_key is already in use"})
        return
    }

    now := time.Now()
    result, err := database.DB.Exec(
        `INSERT INTO sites (name, site_key, domain, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
        req.Name, req.SiteKey, req.Domain, req.Description, now, now,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create site"})
        return
    }

    id, _ := result.LastInsertId()
    site := models.Site{
        ID:          int(id),
        Name:        req.Name,
        SiteKey:     req.SiteKey,
        Domain:      req.Domain,
        Description: req.Description,
        CreatedAt:   now,
        UpdatedAt:   now,
    }

    c.JSON(http.StatusCreated, site)
}

func UpdateSite(c *gin.Context) {
    id := c.Param("id")
    var req models.UpdateSiteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    _, err := database.DB.Exec(
        `UPDATE sites SET name = ?, domain = ?, description = ?, updated_at = ? WHERE id = ?`,
        req.Name, req.Domain, req.Description, time.Now(), id,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update site"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Site updated successfully"})
}

func DeleteSite(c *gin.Context) {
    id := c.Param("id")
    result, err := database.DB.Exec("DELETE FROM sites WHERE id = ?", id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete site"})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Site not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Site deleted successfully"})
}
