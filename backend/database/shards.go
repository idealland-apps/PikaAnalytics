package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"sync"
	"time"

	"pikaanalytics-backend/utils"
)

// ShardKey identifies a monthly visits shard.
type ShardKey struct {
	Year  int
	Month int
}

func (k ShardKey) String() string { return fmt.Sprintf("%04d-%02d", k.Year, k.Month) }

func ShardKeyForTime(t time.Time) ShardKey {
	t = t.UTC()
	return ShardKey{Year: t.Year(), Month: int(t.Month())}
}

func ParseShardKey(s string) (ShardKey, error) {
	var k ShardKey
	if _, err := fmt.Sscanf(s, "%04d-%02d", &k.Year, &k.Month); err != nil {
		return k, fmt.Errorf("invalid month %q (expected YYYY-MM)", s)
	}
	if k.Month < 1 || k.Month > 12 {
		return k, fmt.Errorf("invalid month %q", s)
	}
	return k, nil
}

var (
	shardMu    sync.RWMutex
	shardCache = map[ShardKey]*sql.DB{}
	shardFile  = regexp.MustCompile(`^visits_(\d{4})_(\d{2})\.db$`)
)

const shardSchema = `
CREATE TABLE IF NOT EXISTS page_views (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id INTEGER NOT NULL,
    session_id TEXT,
    path TEXT,
    page_title TEXT,
    referrer TEXT,
    user_agent TEXT,
    ip TEXT,
    country TEXT,
    city TEXT,
    browser TEXT,
    os TEXT,
    device_type TEXT,
    is_bot INTEGER DEFAULT 0,
    screen_width INTEGER,
    screen_height INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_page_views_site_created ON page_views(site_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_page_views_path ON page_views(path);
CREATE INDEX IF NOT EXISTS idx_page_views_session ON page_views(session_id);
`

// InitShards ensures the visits directory exists. Shards are opened lazily.
func InitShards() {
	if err := os.MkdirAll(utils.GetVisitsDir(), 0755); err != nil {
		log.Fatalf("Failed to create visits directory: %v", err)
	}
}

// GetShard opens (or returns the cached) shard for a given key.
func GetShard(key ShardKey) (*sql.DB, error) {
	shardMu.RLock()
	if db, ok := shardCache[key]; ok {
		shardMu.RUnlock()
		return db, nil
	}
	shardMu.RUnlock()

	shardMu.Lock()
	defer shardMu.Unlock()
	if db, ok := shardCache[key]; ok {
		return db, nil
	}

	path := utils.GetVisitsDbPath(key.Year, key.Month)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open shard %s: %w", key, err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping shard %s: %w", key, err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	applyPragmas(db)
	if _, err := db.Exec(shardSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("init shard schema %s: %w", key, err)
	}

	shardCache[key] = db
	log.Printf("Opened visits shard: %s", path)
	return db, nil
}

// GetShardForTime returns the shard that should hold rows with this timestamp.
func GetShardForTime(t time.Time) (*sql.DB, error) {
	return GetShard(ShardKeyForTime(t))
}

// ListShardKeys scans the visits directory and returns existing shards in chronological order.
func ListShardKeys() ([]ShardKey, error) {
	entries, err := os.ReadDir(utils.GetVisitsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var keys []ShardKey
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := shardFile.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		var k ShardKey
		fmt.Sscanf(m[1], "%d", &k.Year)
		fmt.Sscanf(m[2], "%d", &k.Month)
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Year != keys[j].Year {
			return keys[i].Year < keys[j].Year
		}
		return keys[i].Month < keys[j].Month
	})
	return keys, nil
}

// CloseShards closes all open shard handles.
func CloseShards() {
	shardMu.Lock()
	defer shardMu.Unlock()
	for k, db := range shardCache {
		if err := db.Close(); err != nil {
			log.Printf("Error closing shard %s: %v", k, err)
		}
	}
	shardCache = map[ShardKey]*sql.DB{}
}

// PingShards pings all currently open shards.
func PingShards() {
	shardMu.RLock()
	defer shardMu.RUnlock()
	for k, db := range shardCache {
		if err := db.Ping(); err != nil {
			log.Printf("Shard %s health check failed: %v", k, err)
		}
	}
}
