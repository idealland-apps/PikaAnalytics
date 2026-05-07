package handlers

import (
    "database/sql"
    "net/http"

    "pikaanalytics-backend/database"
    "pikaanalytics-backend/middleware"
    "pikaanalytics-backend/models"
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
)

func Login(c *gin.Context) {
    var req models.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var id int
    var username, hashedPassword string
    err := database.DB.QueryRow(
        "SELECT id, username, password FROM users WHERE username = ?",
        req.Username,
    ).Scan(&id, &username, &hashedPassword)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    token, err := middleware.GenerateJWT(id, username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    isDefault := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte("admin123")) == nil

    c.JSON(http.StatusOK, gin.H{
        "token": token,
        "user": gin.H{
            "id":                   id,
            "username":             username,
            "is_default_password":  isDefault,
        },
    })
}

func ChangePassword(c *gin.Context) {
    var req models.ChangePasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userID, _ := c.Get("user_id")

    var hashed string
    if err := database.DB.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&hashed); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(req.CurrentPassword)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
        return
    }

    if len(req.NewPassword) < 6 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be at least 6 characters"})
        return
    }

    newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    if _, err := database.DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(newHash), userID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
