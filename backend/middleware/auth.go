package middleware

import (
    "net/http"
    "strings"
    "pikaanalytics-backend/config"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return config.GetJWTSecret(), nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set("user_id", int(claims["user_id"].(float64)))
            c.Set("username", claims["username"].(string))
        }

        c.Next()
    }
}

func GenerateJWT(userID int, username string) (string, error) {
    claims := jwt.MapClaims{
        "user_id":  userID,
        "username": username,
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(config.GetJWTSecret())
}