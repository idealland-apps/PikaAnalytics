package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"pikaanalytics-backend/database"
)

// resolveSiteFilter parses site_key/site_id and returns a (condition, args) pair
// that can be appended to WHERE clauses. ok=false means a response was already written.
func resolveSiteFilter(c *gin.Context) (string, []interface{}, bool) {
	if siteKey := c.Query("site_key"); siteKey != "" {
		var siteID int
		if err := database.DB.QueryRow("SELECT id FROM sites WHERE site_key = ?", siteKey).Scan(&siteID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site_key"})
			return "", nil, false
		}
		return "site_id = ?", []interface{}{siteID}, true
	}
	if siteIDStr := c.Query("site_id"); siteIDStr != "" {
		siteID, err := strconv.Atoi(siteIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site_id"})
			return "", nil, false
		}
		return "site_id = ?", []interface{}{siteID}, true
	}
	return "", nil, true
}

// resolveMonthShard returns the shard for ?month=YYYY-MM (default: current UTC month).
// Returns (db, key, ok). If the shard file doesn't exist yet it's created on demand.
func resolveMonthShard(c *gin.Context) (*sql.DB, database.ShardKey, bool) {
	var key database.ShardKey
	if monthStr := c.Query("month"); monthStr != "" {
		k, err := database.ParseShardKey(monthStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return nil, key, false
		}
		key = k
	} else {
		key = database.ShardKeyForTime(time.Now().UTC())
	}
	db, err := database.GetShard(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open visits shard"})
		return nil, key, false
	}
	return db, key, true
}

// buildWhere joins a site filter (possibly empty) into a WHERE clause.
func buildWhere(siteCond string) string {
	if siteCond == "" {
		return ""
	}
	return "WHERE " + siteCond
}

func scanTopRows(rows *sql.Rows) ([]gin.H, error) {
	defer rows.Close()
	var results []gin.H
	for rows.Next() {
		var label sql.NullString
		var count int
		if err := rows.Scan(&label, &count); err != nil {
			return nil, err
		}
		text := label.String
		if !label.Valid || text == "" {
			text = "Unknown"
		}
		results = append(results, gin.H{"label": text, "count": count})
	}
	return results, rows.Err()
}

// GetAnalyticsMonths lists available shard months so the frontend can populate the selector.
// The current month is always included even if it has no shard file yet.
func GetAnalyticsMonths(c *gin.Context) {
	keys, err := database.ListShardKeys()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list months"})
		return
	}
	current := database.ShardKeyForTime(time.Now().UTC())
	have := false
	for _, k := range keys {
		if k == current {
			have = true
			break
		}
	}
	if !have {
		keys = append(keys, current)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Year != keys[j].Year {
			return keys[i].Year > keys[j].Year
		}
		return keys[i].Month > keys[j].Month
	})
	months := make([]string, 0, len(keys))
	for _, k := range keys {
		months = append(months, k.String())
	}
	c.JSON(http.StatusOK, gin.H{"months": months, "current": current.String()})
}

func GetAnalyticsOverview(c *gin.Context) {
	siteCond, args, ok := resolveSiteFilter(c)
	if !ok {
		return
	}
	db, _, ok := resolveMonthShard(c)
	if !ok {
		return
	}
	where := buildWhere(siteCond)

	var totalViews, uniquePages, uniqueReferrers, uniqueIPs int
	if err := db.QueryRow(fmt.Sprintf(
		`SELECT COUNT(*), COUNT(DISTINCT path), COUNT(DISTINCT referrer), COUNT(DISTINCT ip) FROM page_views %s`, where),
		args...).Scan(&totalViews, &uniquePages, &uniqueReferrers, &uniqueIPs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch overview metrics"})
		return
	}

	sessionFilter := "session_id IS NOT NULL AND session_id != ''"
	sessWhere := where
	if sessWhere == "" {
		sessWhere = "WHERE " + sessionFilter
	} else {
		sessWhere = sessWhere + " AND " + sessionFilter
	}
	var sessions, bounced int
	var avgDurationSec sql.NullFloat64
	sessQuery := fmt.Sprintf(`
        SELECT COUNT(*) AS sessions,
               COALESCE(SUM(CASE WHEN views = 1 THEN 1 ELSE 0 END), 0) AS bounced,
               AVG(duration_sec) AS avg_duration
        FROM (
            SELECT session_id,
                   COUNT(*) AS views,
                   (julianday(MAX(created_at)) - julianday(MIN(created_at))) * 86400.0 AS duration_sec
            FROM page_views %s
            GROUP BY session_id
        )`, sessWhere)
	if err := db.QueryRow(sessQuery, args...).Scan(&sessions, &bounced, &avgDurationSec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch session metrics"})
		return
	}
	bounceRate := 0.0
	if sessions > 0 {
		bounceRate = float64(bounced) / float64(sessions)
	}
	avgDuration := 0.0
	if avgDurationSec.Valid {
		avgDuration = avgDurationSec.Float64
	}

	browserRows, err := db.Query(fmt.Sprintf(
		`SELECT browser, COUNT(*) FROM page_views %s GROUP BY browser ORDER BY COUNT(*) DESC LIMIT 5`, where), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch browser metrics"})
		return
	}
	browsers, err := scanTopRows(browserRows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse browser metrics"})
		return
	}

	osRows, err := db.Query(fmt.Sprintf(
		`SELECT os, COUNT(*) FROM page_views %s GROUP BY os ORDER BY COUNT(*) DESC LIMIT 5`, where), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch OS metrics"})
		return
	}
	oss, err := scanTopRows(osRows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse OS metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_views":          totalViews,
		"unique_pages":         uniquePages,
		"unique_referrers":     uniqueReferrers,
		"unique_ips":           uniqueIPs,
		"sessions":             sessions,
		"bounce_rate":          bounceRate,
		"avg_session_duration": avgDuration,
		"top_browsers":         browsers,
		"top_os":               oss,
	})
}

// realtimeShards returns the current month shard plus, when within 30 minutes
// of midnight UTC on the 1st, the previous month shard, so the rolling window
// can span a month boundary.
func realtimeShards(now time.Time) ([]*sql.DB, error) {
	cur, err := database.GetShard(database.ShardKeyForTime(now))
	if err != nil {
		return nil, err
	}
	dbs := []*sql.DB{cur}
	if now.Day() == 1 && now.Hour() == 0 && now.Minute() < 30 {
		prevTime := now.AddDate(0, 0, -1)
		if prev, err := database.GetShard(database.ShardKeyForTime(prevTime)); err == nil {
			dbs = append(dbs, prev)
		}
	}
	return dbs, nil
}

func GetAnalyticsRealtime(c *gin.Context) {
	siteCond, args, ok := resolveSiteFilter(c)
	if !ok {
		return
	}

	now := time.Now().UTC()
	dbs, err := realtimeShards(now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open visits shard"})
		return
	}

	build := func(extra string) (string, []interface{}) {
		conds := []string{extra}
		argsCopy := append([]interface{}{}, args...)
		if siteCond != "" {
			conds = append(conds, siteCond)
		}
		return "WHERE " + strings.Join(conds, " AND "), argsCopy
	}

	var activeVisitors, pageviews int
	visitorSet := map[string]struct{}{}
	pageCounts := map[string]int{}
	refCounts := map[string]int{}
	bucketCounts := map[string]int{}

	for _, db := range dbs {
		// Active visitors + pageviews in last 5 minutes.
		w5, a5 := build("created_at >= datetime('now', '-5 minutes')")
		idRows, err := db.Query(fmt.Sprintf(
			`SELECT COALESCE(NULLIF(session_id, ''), ip) FROM page_views %s`, w5), a5...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch realtime visitors"})
			return
		}
		for idRows.Next() {
			var id sql.NullString
			if err := idRows.Scan(&id); err != nil {
				idRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse realtime visitors"})
				return
			}
			pageviews++
			if id.Valid && id.String != "" {
				visitorSet[id.String] = struct{}{}
			}
		}
		idRows.Close()

		pagesRows, err := db.Query(fmt.Sprintf(
			`SELECT COALESCE(path, ''), COUNT(*) FROM page_views %s GROUP BY path`, w5), a5...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch realtime pages"})
			return
		}
		mergeCounts(pagesRows, pageCounts, c)
		refRows, err := db.Query(fmt.Sprintf(
			`SELECT COALESCE(referrer, ''), COUNT(*) FROM page_views %s GROUP BY referrer`, w5), a5...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch realtime referrers"})
			return
		}
		mergeCounts(refRows, refCounts, c)

		w30, a30 := build("created_at >= datetime('now', '-30 minutes')")
		bRows, err := db.Query(fmt.Sprintf(
			`SELECT substr(created_at, 1, 10) || 'T' || substr(created_at, 12, 5) || ':00Z' AS minute, COUNT(*)
             FROM page_views %s GROUP BY minute`, w30), a30...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch realtime buckets"})
			return
		}
		mergeCounts(bRows, bucketCounts, c)
	}
	activeVisitors = len(visitorSet)

	c.JSON(http.StatusOK, gin.H{
		"active_visitors": activeVisitors,
		"pageviews":       pageviews,
		"top_pages":       topNFromMap(pageCounts, 10),
		"top_referrers":   topNFromMap(refCounts, 10),
		"buckets":         bucketsFromMap(bucketCounts),
	})
}

func mergeCounts(rows *sql.Rows, dst map[string]int, c *gin.Context) {
	defer rows.Close()
	for rows.Next() {
		var key sql.NullString
		var n int
		if err := rows.Scan(&key, &n); err != nil {
			return
		}
		dst[key.String] += n
	}
}

func topNFromMap(m map[string]int, n int) []gin.H {
	type kv struct {
		k string
		v int
	}
	list := make([]kv, 0, len(m))
	for k, v := range m {
		list = append(list, kv{k, v})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].v > list[j].v })
	if len(list) > n {
		list = list[:n]
	}
	out := make([]gin.H, 0, len(list))
	for _, it := range list {
		label := it.k
		if label == "" {
			label = "Unknown"
		}
		out = append(out, gin.H{"label": label, "count": it.v})
	}
	return out
}

func bucketsFromMap(m map[string]int) []gin.H {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]gin.H, 0, len(keys))
	for _, k := range keys {
		out = append(out, gin.H{"minute": k, "views": m[k]})
	}
	return out
}

func topByQuery(c *gin.Context, column, jsonKey string) {
	siteCond, args, ok := resolveSiteFilter(c)
	if !ok {
		return
	}
	db, _, ok := resolveMonthShard(c)
	if !ok {
		return
	}
	where := buildWhere(siteCond)
	rows, err := db.Query(fmt.Sprintf(
		`SELECT %s, COUNT(*) FROM page_views %s GROUP BY %s ORDER BY COUNT(*) DESC LIMIT 20`, column, where, column), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch metrics"})
		return
	}
	items, err := scanTopRows(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse metrics"})
		return
	}
	c.JSON(http.StatusOK, gin.H{jsonKey: items})
}

func GetAnalyticsPages(c *gin.Context)     { topByQuery(c, "path", "pages") }
func GetAnalyticsReferrers(c *gin.Context) { topByQuery(c, "referrer", "referrers") }
func GetAnalyticsLocations(c *gin.Context) { topByQuery(c, "country", "locations") }

func GetAnalyticsDevices(c *gin.Context) {
	siteCond, args, ok := resolveSiteFilter(c)
	if !ok {
		return
	}
	db, _, ok := resolveMonthShard(c)
	if !ok {
		return
	}
	where := buildWhere(siteCond)

	browserRows, err := db.Query(fmt.Sprintf(
		`SELECT browser, COUNT(*) FROM page_views %s GROUP BY browser ORDER BY COUNT(*) DESC LIMIT 20`, where), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch browser metrics"})
		return
	}
	browsers, err := scanTopRows(browserRows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse browser metrics"})
		return
	}

	osRows, err := db.Query(fmt.Sprintf(
		`SELECT os, COUNT(*) FROM page_views %s GROUP BY os ORDER BY COUNT(*) DESC LIMIT 20`, where), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch OS metrics"})
		return
	}
	oss, err := scanTopRows(osRows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse OS metrics"})
		return
	}

	deviceRows, err := db.Query(fmt.Sprintf(
		`SELECT device_type, COUNT(*) FROM page_views %s GROUP BY device_type ORDER BY COUNT(*) DESC LIMIT 20`, where), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch device metrics"})
		return
	}
	devices, err := scanTopRows(deviceRows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse device metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"browsers": browsers, "os": oss, "devices": devices})
}

func GetAnalyticsVisits(c *gin.Context) {
	siteCond, args, ok := resolveSiteFilter(c)
	if !ok {
		return
	}
	db, _, ok := resolveMonthShard(c)
	if !ok {
		return
	}
	where := buildWhere(siteCond)

	rows, err := db.Query(fmt.Sprintf(
		`SELECT substr(created_at, 1, 10) AS day, COUNT(*), COUNT(DISTINCT ip)
         FROM page_views %s
         GROUP BY day ORDER BY day ASC`, where), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch visit metrics"})
		return
	}
	defer rows.Close()

	var visits []gin.H
	for rows.Next() {
		var day sql.NullString
		var pv, uv int
		if err := rows.Scan(&day, &pv, &uv); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to parse visit metrics: %v", err)})
			return
		}
		visits = append(visits, gin.H{"date": day.String, "views": pv, "uniques": uv})
	}
	c.JSON(http.StatusOK, gin.H{"visits": visits})
}

func GetAnalyticsRecent(c *gin.Context) {
	siteCond, args, ok := resolveSiteFilter(c)
	if !ok {
		return
	}
	db, _, ok := resolveMonthShard(c)
	if !ok {
		return
	}
	where := buildWhere(siteCond)

	limit := 100
	if s := c.Query("limit"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	// sites lives in main DB; load a small map and enrich rows in Go.
	siteRows, err := database.DB.Query(`SELECT id, name, site_key FROM sites`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load sites"})
		return
	}
	type siteInfo struct{ name, key string }
	siteMap := map[int]siteInfo{}
	for siteRows.Next() {
		var id int
		var name, key string
		if err := siteRows.Scan(&id, &name, &key); err == nil {
			siteMap[id] = siteInfo{name, key}
		}
	}
	siteRows.Close()

	query := fmt.Sprintf(
		`SELECT id, site_id, COALESCE(path, ''), COALESCE(page_title, ''), COALESCE(referrer, ''),
                COALESCE(ip, ''), COALESCE(country, ''), COALESCE(city, ''),
                COALESCE(browser, ''), COALESCE(os, ''), COALESCE(device_type, ''),
                is_bot, created_at
         FROM page_views %s ORDER BY created_at DESC LIMIT %d`, where, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent visits"})
		return
	}
	defer rows.Close()

	var visits []gin.H
	for rows.Next() {
		var id, siteID, isBot int
		var path, title, referrer, ip, country, city, browser, osName, device string
		var createdAt string
		if err := rows.Scan(&id, &siteID, &path, &title, &referrer, &ip, &country, &city, &browser, &osName, &device, &isBot, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse recent visits"})
			return
		}
		info := siteMap[siteID]
		visits = append(visits, gin.H{
			"id":          id,
			"site_id":     siteID,
			"site_name":   info.name,
			"site_key":    info.key,
			"path":        path,
			"page_title":  title,
			"referrer":    referrer,
			"ip":          ip,
			"country":     country,
			"city":        city,
			"browser":     browser,
			"os":          osName,
			"device_type": device,
			"is_bot":      isBot == 1,
			"created_at":  createdAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"visits": visits})
}
