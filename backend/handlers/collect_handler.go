package handlers

import (
    "fmt"
    "net/http"
    "time"

    "pikaanalytics-backend/database"
    "pikaanalytics-backend/logging"
    "pikaanalytics-backend/models"
    "github.com/gin-gonic/gin"
)

func TrackScript(c *gin.Context) {
    siteKey := c.Query("site")
    if siteKey == "" {
        c.String(http.StatusBadRequest, "/* missing site parameter */")
        return
    }

    script := `;(function(){
  try {
    var siteKey = %q;
    var scriptUrl = (document.currentScript && document.currentScript.src) ? document.currentScript.src : window.location.href;
    var endpoint = new URL('/api/pulse', scriptUrl).toString();

    function sessionId() {
      try {
        var k = 'pa_sid';
        var sid = sessionStorage.getItem(k);
        if (!sid) {
          sid = (crypto && crypto.randomUUID) ? crypto.randomUUID()
            : (Date.now().toString(36) + Math.random().toString(36).slice(2));
          sessionStorage.setItem(k, sid);
        }
        return sid;
      } catch (e) { return ''; }
    }

    function send() {
      var payload = {
        site_key: siteKey,
        session_id: sessionId(),
        path: window.location.pathname + window.location.search,
        page_title: document.title || '',
        referrer: document.referrer || '',
        screen_width: window.screen ? window.screen.width : 0,
        screen_height: window.screen ? window.screen.height : 0
      };
      var body = JSON.stringify(payload);
      fetch(endpoint, {
        method: 'POST',
        body: body,
        headers: { 'Content-Type': 'application/json' },
        keepalive: true,
        credentials: 'omit'
      }).catch(function(){});
    }

    if (document.readyState === 'complete' || document.readyState === 'interactive') {
      send();
    } else {
      document.addEventListener('DOMContentLoaded', send, { once: true });
    }
  } catch (e) {
    if (window.console && console.error) console.error('PikaAnalytics track error', e);
  }
})();`

    c.Header("Cache-Control", "public, max-age=300")
    c.Data(http.StatusOK, "application/javascript; charset=utf-8", []byte(fmt.Sprintf(script, siteKey)))
}

func CollectEvent(c *gin.Context) {
    var req models.CollectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var siteID int
    if err := database.DB.QueryRow("SELECT id FROM sites WHERE site_key = ?", req.SiteKey).Scan(&siteID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site_key"})
        return
    }

    if req.IP == "" {
        req.IP = c.ClientIP()
    }
    userAgent := req.UserAgent
    if userAgent == "" {
        userAgent = c.GetHeader("User-Agent")
    }
    if req.Referrer == "" {
        req.Referrer = c.GetHeader("Referer")
    }

    uaInfo := logging.ParseUserAgent(userAgent)
    ipInfo := logging.ParseIPAddress(req.IP)

    browser := req.Browser
    if browser == "" {
        browser = uaInfo.Browser
    }
    osName := req.OS
    if osName == "" {
        osName = uaInfo.OS
    }
    deviceType := req.DeviceType
    if deviceType == "" {
        deviceType = uaInfo.DeviceType
    }
    isBot := req.IsBot || uaInfo.IsBot
    country := req.Country
    if country == "" {
        country = ipInfo.Country
    }
    city := req.City
    if city == "" {
        city = ipInfo.City
    }

    now := time.Now().UTC()
    shard, err := database.GetShardForTime(now)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open visits shard"})
        return
    }
    _, err = shard.Exec(
        `INSERT INTO page_views (site_id, session_id, path, page_title, referrer, user_agent, ip, country, city, browser, os, device_type, is_bot, screen_width, screen_height, created_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        siteID,
        req.SessionID,
        req.Path,
        req.PageTitle,
        req.Referrer,
        userAgent,
        req.IP,
        country,
        city,
        browser,
        osName,
        deviceType,
        boolToInt(isBot),
        req.ScreenWidth,
        req.ScreenHeight,
        now,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save analytics event"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func boolToInt(value bool) int {
    if value {
        return 1
    }
    return 0
}
