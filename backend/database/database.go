package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"

	"golang.org/x/crypto/bcrypt"
	"pikaanalytics-backend/utils"
)

var DB *sql.DB

func InitDB() {
	dataPath := utils.GetDataPath()
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}
	if err := os.MkdirAll(utils.GetVisitsDir(), 0755); err != nil {
		log.Fatal("Failed to create visits directory:", err)
	}

	dbPath := utils.GetMainDbPath()
	log.Printf("Initializing main database at: %s", dbPath)

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	applyPragmas(DB)
	createTables()
	InitShards()
	log.Println("Database connected, optimized, and tables created")
}

func applyPragmas(db *sql.DB) {
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=1000",
		"PRAGMA temp_store=MEMORY",
	} {
		if _, err := db.Exec(pragma); err != nil {
			log.Printf("Warning: %s failed: %v", pragma, err)
		}
	}
}

func createTables() {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS system_config (
            config_key TEXT PRIMARY KEY,
            config_value TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );`,
		`CREATE TABLE IF NOT EXISTS sites (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            site_key TEXT UNIQUE NOT NULL,
            domain TEXT,
            description TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );`,
		`CREATE TABLE IF NOT EXISTS site_configs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            site_id INTEGER NOT NULL,
            config_key TEXT NOT NULL,
            config_value TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (site_id) REFERENCES sites (id) ON DELETE CASCADE
        );`,
	}

	for _, stmt := range statements {
		if _, err := DB.Exec(stmt); err != nil {
			log.Fatalf("Failed to execute schema statement: %v\n%s", err, stmt)
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}
	DB.Exec("INSERT OR IGNORE INTO users (username, password) VALUES (?, ?)", "admin", string(hashedPassword))
}

func Close() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing main database: %v", err)
		}
	}
	CloseShards()
}
