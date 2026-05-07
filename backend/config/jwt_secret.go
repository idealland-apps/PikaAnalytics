package config

import (
    "crypto/rand"
    "database/sql"
    "encoding/base64"
    "fmt"
    "log"
    "sync"
)

var (
    jwtSecret []byte
    secretMutex sync.RWMutex
)

// GenerateRandomSecret creates a cryptographically secure random 32-byte secret
func GenerateRandomSecret() ([]byte, error) {
    secret := make([]byte, 32)
    _, err := rand.Read(secret)
    if err != nil {
        return nil, fmt.Errorf("failed to generate random secret: %w", err)
    }
    return secret, nil
}

// InitJWTSecret initializes the JWT secret on application startup
// It retrieves an existing secret from database or generates a new one if none exists
func InitJWTSecret(db *sql.DB) error {
    secretMutex.Lock()
    defer secretMutex.Unlock()

    // Try to retrieve existing secret from database
    secret, err := getSecretFromDB(db)
    if err != nil {
        return fmt.Errorf("failed to retrieve JWT secret from database: %w", err)
    }

    if secret != nil {
        jwtSecret = secret
        log.Println("JWT secret loaded from database")
        return nil
    }

    // No secret exists, generate a new one
    newSecret, err := GenerateRandomSecret()
    if err != nil {
        return fmt.Errorf("failed to generate new JWT secret: %w", err)
    }

    // Store the new secret in database
    if err := storeSecretInDB(db, newSecret); err != nil {
        return fmt.Errorf("failed to store JWT secret in database: %w", err)
    }

    jwtSecret = newSecret
    log.Println("New JWT secret generated and stored in database")
    return nil
}

// GetJWTSecret returns the current JWT secret (thread-safe)
func GetJWTSecret() []byte {
    secretMutex.RLock()
    defer secretMutex.RUnlock()
    
    if jwtSecret == nil {
        log.Fatal("JWT secret not initialized. Call InitJWTSecret first.")
    }
    
    // Return a copy to prevent external modification
    secret := make([]byte, len(jwtSecret))
    copy(secret, jwtSecret)
    return secret
}

// getSecretFromDB retrieves the JWT secret from the database
func getSecretFromDB(db *sql.DB) ([]byte, error) {
    var secretBase64 string
    err := db.QueryRow("SELECT config_value FROM system_config WHERE config_key = ?", "jwt_secret").Scan(&secretBase64)
    
    if err == sql.ErrNoRows {
        return nil, nil // No secret exists yet
    }
    if err != nil {
        return nil, err
    }

    // Decode base64 secret
    secret, err := base64.StdEncoding.DecodeString(secretBase64)
    if err != nil {
        return nil, fmt.Errorf("failed to decode JWT secret: %w", err)
    }

    return secret, nil
}

// storeSecretInDB stores the JWT secret in the database
func storeSecretInDB(db *sql.DB, secret []byte) error {
    // Encode secret as base64 for storage
    secretBase64 := base64.StdEncoding.EncodeToString(secret)
    
    _, err := db.Exec(
        "INSERT OR REPLACE INTO system_config (config_key, config_value) VALUES (?, ?)",
        "jwt_secret", secretBase64,
    )
    return err
}