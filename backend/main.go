package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"
    "pikaanalytics-backend/config"
    "pikaanalytics-backend/database"
    "pikaanalytics-backend/handlers"
    "pikaanalytics-backend/middleware"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
)

func main() {
    database.InitDB()

    if err := config.InitJWTSecret(database.DB); err != nil {
        log.Fatal("Failed to initialize JWT secret:", err)
    }

    r := gin.Default()

    publicCORS := cors.New(cors.Config{
        AllowAllOrigins:  true,
        AllowMethods:     []string{"GET", "POST", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: false,
        MaxAge:           24 * time.Hour,
    })

    corsOrigins := os.Getenv("CORS_ORIGINS")
    var adminCORSConfig cors.Config
    if corsOrigins == "" {
        adminCORSConfig = cors.Config{
            AllowAllOrigins:  true,
            AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
            ExposeHeaders:    []string{"Content-Length"},
            AllowCredentials: false,
        }
    } else {
        allowedOrigins := strings.Split(corsOrigins, ",")
        for i, origin := range allowedOrigins {
            allowedOrigins[i] = strings.TrimSpace(origin)
        }
        adminCORSConfig = cors.Config{
            AllowOrigins:     allowedOrigins,
            AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
            ExposeHeaders:    []string{"Content-Length"},
            AllowCredentials: true,
        }
    }
    adminCORS := cors.New(adminCORSConfig)

    // Public routes (open CORS — embedded on third-party sites)
    r.GET("/pulse.js", publicCORS, handlers.TrackScript)
    r.POST("/api/pulse", publicCORS, handlers.CollectEvent)
    r.OPTIONS("/api/pulse", publicCORS)

    // Login uses admin CORS (whitelist when configured)
    r.POST("/api/login", adminCORS, handlers.Login)

    // Protected routes
    api := r.Group("/api")
    api.Use(adminCORS, middleware.AuthMiddleware())
    {
        api.GET("/sites", handlers.GetSites)
        api.GET("/sites/:id", handlers.GetSite)
        api.POST("/sites", handlers.CreateSite)
        api.PUT("/sites/:id", handlers.UpdateSite)
        api.DELETE("/sites/:id", handlers.DeleteSite)

        api.GET("/analytics/overview", handlers.GetAnalyticsOverview)
        api.GET("/analytics/pages", handlers.GetAnalyticsPages)
        api.GET("/analytics/referrers", handlers.GetAnalyticsReferrers)
        api.GET("/analytics/devices", handlers.GetAnalyticsDevices)
        api.GET("/analytics/locations", handlers.GetAnalyticsLocations)
        api.GET("/analytics/visits", handlers.GetAnalyticsVisits)
        api.GET("/analytics/recent", handlers.GetAnalyticsRecent)
        api.GET("/analytics/realtime", handlers.GetAnalyticsRealtime)
        api.GET("/analytics/months", handlers.GetAnalyticsMonths)

        api.POST("/change-password", handlers.ChangePassword)
        api.GET("/config/:key", handlers.GetConfig)
        api.PUT("/config/:key", handlers.UpdateConfig)
        api.GET("/version", handlers.GetVersion)
    }

    // Serve admin frontend at /admin
    r.Static("/admin", "./frontend/")

    // Root redirects to admin login
    r.GET("/", func(c *gin.Context) {
        c.Redirect(http.StatusFound, "/admin/")
    })

    r.NoRoute(func(c *gin.Context) {
        path := c.Request.URL.Path
        if strings.HasPrefix(path, "/admin") {
            c.File("./frontend/index.html")
            return
        }
        c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
    })

    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        for range ticker.C {
            if err := database.DB.Ping(); err != nil {
                log.Printf("Main database health check failed: %v", err)
            }
            database.PingShards()
        }
    }()

    srv := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }

    go func() {
        log.Println("PikaAnalytics server starting on :8080")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")

    if err := database.DB.Close(); err != nil {
        log.Printf("Error closing main database: %v", err)
    }
    database.CloseShards()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    log.Println("Server exited")
}
