package logging

import (
	"sync"
	"time"
)

// UserAgentCache provides in-memory caching for parsed user agent strings
type UserAgentCache struct {
	cache map[string]CachedUserAgentInfo
	mutex sync.RWMutex
	maxSize int
	ttl   time.Duration
}

// CachedUserAgentInfo includes the parsed info and timestamp
type CachedUserAgentInfo struct {
	Info      UserAgentInfo
	Timestamp time.Time
}

var (
	uaCache     *UserAgentCache
	uaCacheOnce sync.Once
)

// GetUserAgentCache returns the singleton user agent cache instance
func GetUserAgentCache() *UserAgentCache {
	uaCacheOnce.Do(func() {
		uaCache = &UserAgentCache{
			cache:   make(map[string]CachedUserAgentInfo),
			maxSize: 1000,              // Cache up to 1000 unique user agents
			ttl:     24 * time.Hour,    // Cache for 24 hours
		}
		
		// Start cleanup goroutine
		go uaCache.cleanupWorker()
	})
	return uaCache
}

// cleanupWorker periodically removes expired entries from cache
func (uac *UserAgentCache) cleanupWorker() {
	ticker := time.NewTicker(1 * time.Hour) // Cleanup every hour
	defer ticker.Stop()

	for range ticker.C {
		uac.cleanup()
	}
}

// cleanup removes expired entries and enforces max size
func (uac *UserAgentCache) cleanup() {
	uac.mutex.Lock()
	defer uac.mutex.Unlock()

	now := time.Now()
	
	// Remove expired entries
	for userAgent, cachedInfo := range uac.cache {
		if now.Sub(cachedInfo.Timestamp) > uac.ttl {
			delete(uac.cache, userAgent)
		}
	}

	// If still over max size, remove oldest entries
	if len(uac.cache) > uac.maxSize {
		type entry struct {
			userAgent string
			timestamp time.Time
		}

		// Collect all entries with timestamps
		entries := make([]entry, 0, len(uac.cache))
		for userAgent, cachedInfo := range uac.cache {
			entries = append(entries, entry{
				userAgent: userAgent,
				timestamp: cachedInfo.Timestamp,
			})
		}

		// Sort by timestamp (oldest first) and remove excess
		for i := 0; i < len(entries)-uac.maxSize; i++ {
			delete(uac.cache, entries[i].userAgent)
		}
	}
}

// Get retrieves user agent info from cache or parses it if not cached
func (uac *UserAgentCache) Get(userAgent string) UserAgentInfo {
	// Check cache first
	uac.mutex.RLock()
	if cachedInfo, exists := uac.cache[userAgent]; exists {
		if time.Since(cachedInfo.Timestamp) <= uac.ttl {
			uac.mutex.RUnlock()
			return cachedInfo.Info
		}
	}
	uac.mutex.RUnlock()

	// Parse user agent (this is expensive)
	info := parseUserAgentUncached(userAgent)

	// Cache the result
	uac.mutex.Lock()
	// Check size limit before adding
	if len(uac.cache) >= uac.maxSize {
		// Remove one random old entry (simple eviction)
		for oldUserAgent := range uac.cache {
			delete(uac.cache, oldUserAgent)
			break
		}
	}
	
	uac.cache[userAgent] = CachedUserAgentInfo{
		Info:      info,
		Timestamp: time.Now(),
	}
	uac.mutex.Unlock()

	return info
}

// ParseUserAgent is the new cached version of the original ParseUserAgent function
func ParseUserAgent(userAgent string) UserAgentInfo {
	cache := GetUserAgentCache()
	return cache.Get(userAgent)
}

// ClearCache clears the user agent cache (useful for testing)
func (uac *UserAgentCache) ClearCache() {
	uac.mutex.Lock()
	defer uac.mutex.Unlock()
	uac.cache = make(map[string]CachedUserAgentInfo)
}

// GetCacheStats returns cache statistics
func (uac *UserAgentCache) GetCacheStats() (size int, maxSize int, ttl time.Duration) {
	uac.mutex.RLock()
	defer uac.mutex.RUnlock()
	return len(uac.cache), uac.maxSize, uac.ttl
}