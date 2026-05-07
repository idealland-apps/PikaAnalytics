package models

type CollectRequest struct {
    SiteKey      string `json:"site_key" binding:"required"`
    SessionID    string `json:"session_id"`
    Path         string `json:"path" binding:"required"`
    PageTitle    string `json:"page_title"`
    Referrer     string `json:"referrer"`
    UserAgent    string `json:"user_agent"`
    IP           string `json:"ip"`
    Country      string `json:"country"`
    City         string `json:"city"`
    Browser      string `json:"browser"`
    OS           string `json:"os"`
    DeviceType   string `json:"device_type"`
    IsBot        bool   `json:"is_bot"`
    ScreenWidth  int    `json:"screen_width"`
    ScreenHeight int    `json:"screen_height"`
}
