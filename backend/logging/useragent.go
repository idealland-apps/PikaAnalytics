package logging

import (
	"regexp"
	"strings"
)

// UserAgentInfo contains parsed information from a User-Agent string
type UserAgentInfo struct {
	Browser        string
	BrowserVersion string
	OS             string
	DeviceType     string
	IsBot          bool
}

// Common bot patterns - ordered by specificity
var botPatterns = []string{
	"Googlebot", "Bingbot", "Slurp", "DuckDuckBot", "Baiduspider",
	"YandexBot", "facebookexternalhit", "Twitterbot", "LinkedInBot",
	"WhatsApp", "Applebot", "crawler", "spider", "bot", "Bot",
	"Crawl", "scraper", "Scraper", "fetch", "Fetch", "monitor",
	"Monitor", "check", "Check", "test", "Test", "scan", "Scan",
	"curl", "wget", "python", "Python", "Go-http-client", "Java",
	"PostmanRuntime", "insomnia", "HTTPie", "axios", "requests",
}

// Browser patterns with versions
var browserPatterns = []struct {
	name    string
	pattern *regexp.Regexp
	version *regexp.Regexp
}{
	{"Chrome", regexp.MustCompile(`Chrome/(\d+\.\d+)`), regexp.MustCompile(`Chrome/(\d+\.\d+)`)},
	{"Firefox", regexp.MustCompile(`Firefox/(\d+\.\d+)`), regexp.MustCompile(`Firefox/(\d+\.\d+)`)},
	{"Safari", regexp.MustCompile(`Safari/(\d+)`), regexp.MustCompile(`Version/(\d+\.\d+).*Safari`)},
	{"Edge", regexp.MustCompile(`Edg/(\d+\.\d+)`), regexp.MustCompile(`Edg/(\d+\.\d+)`)},
	{"Opera", regexp.MustCompile(`OPR/(\d+\.\d+)`), regexp.MustCompile(`OPR/(\d+\.\d+)`)},
	{"Internet Explorer", regexp.MustCompile(`MSIE (\d+\.\d+)`), regexp.MustCompile(`MSIE (\d+\.\d+)`)},
}

// OS patterns
var osPatterns = []struct {
	name    string
	pattern *regexp.Regexp
}{
	{"Windows 11", regexp.MustCompile(`Windows NT 10\.0.*(?:; Win64; x64|WOW64)`)},
	{"Windows 10", regexp.MustCompile(`Windows NT 10\.0`)},
	{"Windows 8.1", regexp.MustCompile(`Windows NT 6\.3`)},
	{"Windows 8", regexp.MustCompile(`Windows NT 6\.2`)},
	{"Windows 7", regexp.MustCompile(`Windows NT 6\.1`)},
	{"Windows", regexp.MustCompile(`Windows NT`)},
	{"macOS", regexp.MustCompile(`Mac OS X|Macintosh`)},
	{"iOS", regexp.MustCompile(`iPhone|iPad|iPod`)},
	{"Android", regexp.MustCompile(`Android`)},
	{"Linux", regexp.MustCompile(`Linux`)},
	{"ChromeOS", regexp.MustCompile(`CrOS`)},
	{"Ubuntu", regexp.MustCompile(`Ubuntu`)},
}

// Device type patterns
var devicePatterns = []struct {
	deviceType string
	pattern    *regexp.Regexp
}{
	{"Mobile", regexp.MustCompile(`Mobile|iPhone|Android.*Mobile|BlackBerry|Opera Mini`)},
	{"Tablet", regexp.MustCompile(`iPad|Tablet`)}, // Simplified Android tablet detection will be handled in parseDeviceType
	{"TV", regexp.MustCompile(`TV|Television|SmartTV|WebOS|Tizen`)},
	{"Console", regexp.MustCompile(`PlayStation|Xbox|Nintendo`)},
	{"Desktop", regexp.MustCompile(`.+`)}, // Default fallback
}

// parseUserAgentUncached analyzes a User-Agent string and returns structured information
func parseUserAgentUncached(userAgent string) UserAgentInfo {
	if userAgent == "" || userAgent == "-" {
		return UserAgentInfo{
			Browser:        "Unknown",
			BrowserVersion: "Unknown",
			OS:             "Unknown",
			DeviceType:     "Unknown",
			IsBot:          false,
		}
	}

	info := UserAgentInfo{
		Browser:        "Unknown",
		BrowserVersion: "Unknown",
		OS:             "Unknown",
		DeviceType:     "Desktop", // Default to desktop
		IsBot:          false,
	}

	// Check if it's a bot first
	info.IsBot = isBot(userAgent)

	// If it's a bot, try to identify the bot type as "browser"
	if info.IsBot {
		info.Browser = identifyBot(userAgent)
		info.DeviceType = "Bot"
	} else {
		// Parse browser and version
		info.Browser, info.BrowserVersion = parseBrowser(userAgent)
		
		// Parse device type
		info.DeviceType = parseDeviceType(userAgent)
	}

	// Parse OS
	info.OS = parseOS(userAgent)

	return info
}

// isBot checks if the User-Agent string indicates a bot/crawler
func isBot(userAgent string) bool {
	userAgentLower := strings.ToLower(userAgent)
	
	for _, pattern := range botPatterns {
		if strings.Contains(userAgentLower, strings.ToLower(pattern)) {
			return true
		}
	}
	
	// Additional heuristics for bots
	// Very short or very long user agents are often bots
	if len(userAgent) < 10 || len(userAgent) > 1000 {
		return true
	}
	
	// Check for missing common browser indicators
	if !strings.Contains(userAgent, "Mozilla") && 
	   !strings.Contains(userAgent, "Chrome") && 
	   !strings.Contains(userAgent, "Firefox") && 
	   !strings.Contains(userAgent, "Safari") {
		return true
	}
	
	return false
}

// identifyBot tries to identify the specific bot type
func identifyBot(userAgent string) string {
	userAgentLower := strings.ToLower(userAgent)
	
	// Check for specific known bots
	botMap := map[string]string{
		"googlebot":             "Googlebot",
		"bingbot":              "Bingbot",
		"slurp":                "Yahoo Slurp",
		"duckduckbot":          "DuckDuckBot",
		"baiduspider":          "Baidu Spider",
		"yandexbot":            "YandexBot",
		"facebookexternalhit":  "Facebook Crawler",
		"twitterbot":           "TwitterBot",
		"linkedinbot":          "LinkedInBot",
		"whatsapp":             "WhatsApp",
		"applebot":             "AppleBot",
		"curl":                 "cURL",
		"wget":                 "wget",
		"python":               "Python Client",
		"go-http-client":       "Go HTTP Client",
		"postmanruntime":       "Postman",
		"insomnia":             "Insomnia",
		"httpie":               "HTTPie",
	}
	
	for key, name := range botMap {
		if strings.Contains(userAgentLower, key) {
			return name
		}
	}
	
	// Generic classification
	if strings.Contains(userAgentLower, "crawler") || strings.Contains(userAgentLower, "spider") {
		return "Web Crawler"
	}
	if strings.Contains(userAgentLower, "monitor") || strings.Contains(userAgentLower, "check") {
		return "Monitor"
	}
	if strings.Contains(userAgentLower, "scan") {
		return "Scanner"
	}
	
	return "Unknown Bot"
}

// parseBrowser extracts browser name and version from User-Agent
func parseBrowser(userAgent string) (string, string) {
	// Special case for Safari (needs to be checked before Chrome)
	if strings.Contains(userAgent, "Safari") && !strings.Contains(userAgent, "Chrome") {
		if match := regexp.MustCompile(`Version/(\d+\.\d+).*Safari`).FindStringSubmatch(userAgent); len(match) > 1 {
			return "Safari", match[1]
		}
		return "Safari", "Unknown"
	}
	
	// Check other browsers
	for _, browser := range browserPatterns {
		if browser.pattern.MatchString(userAgent) {
			if match := browser.version.FindStringSubmatch(userAgent); len(match) > 1 {
				return browser.name, match[1]
			}
			return browser.name, "Unknown"
		}
	}
	
	return "Unknown", "Unknown"
}

// parseOS extracts operating system information from User-Agent
func parseOS(userAgent string) string {
	for _, os := range osPatterns {
		if os.pattern.MatchString(userAgent) {
			return os.name
		}
	}
	return "Unknown"
}

// parseDeviceType determines the device type from User-Agent
func parseDeviceType(userAgent string) string {
	// Check for mobile first (has higher priority)
	mobilePattern := regexp.MustCompile(`Mobile|iPhone|Android.*Mobile|BlackBerry|Opera Mini`)
	if mobilePattern.MatchString(userAgent) {
		return "Mobile"
	}
	
	// Check for tablets (including Android tablets without "Mobile")
	tabletPattern := regexp.MustCompile(`iPad|Tablet`)
	androidPattern := regexp.MustCompile(`Android`)
	if tabletPattern.MatchString(userAgent) || androidPattern.MatchString(userAgent) {
		return "Tablet"
	}
	
	// Check other device types
	for _, device := range devicePatterns {
		if device.deviceType != "Mobile" && device.deviceType != "Tablet" && device.deviceType != "Desktop" {
			if device.pattern.MatchString(userAgent) {
				return device.deviceType
			}
		}
	}
	
	return "Desktop" // Default fallback
}