package logging

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

// IPInfo contains parsed information from IP address
type IPInfo struct {
	Country string
	City    string
	ASN     string
}

// IPParser handles IP geolocation parsing using GeoLite2 databases
type IPParser struct {
	cityDB    *geoip2.Reader
	asnDB     *geoip2.Reader
	cityMutex sync.RWMutex
	asnMutex  sync.RWMutex
}

var (
	ipParser     *IPParser
	ipParserOnce sync.Once
)

// GetIPParser returns the singleton IP parser instance
func GetIPParser() *IPParser {
	ipParserOnce.Do(func() {
		ipParser = &IPParser{}
		ipParser.initialize()
	})
	return ipParser
}

// initialize sets up the GeoIP databases
func (p *IPParser) initialize() {
	// Get GeoIP database path from environment variable
	geoIPPath := os.Getenv("GEO_IP_DB_PATH")
	if geoIPPath == "" {
		// Fallback to current directory
		geoIPPath = "."
	}

	// Try to load City database
	cityDBPath := filepath.Join(geoIPPath, "GeoLite2-City.mmdb")
	if _, err := os.Stat(cityDBPath); err == nil {
		if cityDB, err := geoip2.Open(cityDBPath); err == nil {
			p.cityMutex.Lock()
			p.cityDB = cityDB
			p.cityMutex.Unlock()
			log.Printf("Loaded GeoLite2-City database from: %s", cityDBPath)
		} else {
			log.Printf("Failed to open GeoLite2-City database: %v", err)
		}
	} else {
		log.Printf("GeoLite2-City database not found at: %s", cityDBPath)
	}

	// Try to load ASN database
	asnDBPath := filepath.Join(geoIPPath, "GeoLite2-ASN.mmdb")
	if _, err := os.Stat(asnDBPath); err == nil {
		if asnDB, err := geoip2.Open(asnDBPath); err == nil {
			p.asnMutex.Lock()
			p.asnDB = asnDB
			p.asnMutex.Unlock()
			log.Printf("Loaded GeoLite2-ASN database from: %s", asnDBPath)
		} else {
			log.Printf("Failed to open GeoLite2-ASN database: %v", err)
		}
	} else {
		log.Printf("GeoLite2-ASN database not found at: %s", asnDBPath)
	}
}

// ParseIP parses an IP address and returns geolocation information
func (p *IPParser) ParseIP(ipStr string) IPInfo {
	info := IPInfo{
		Country: "",
		City:    "",
		ASN:     "",
	}

	// Parse IP address
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return info
	}

	// Skip private/local IP addresses
	if isPrivateIP(ip) {
		return info
	}

	// Get country and city information
	p.cityMutex.RLock()
	cityDB := p.cityDB
	p.cityMutex.RUnlock()

	if cityDB != nil {
		if record, err := cityDB.City(ip); err == nil {
			if record.Country.Names != nil {
				if countryName, exists := record.Country.Names["en"]; exists {
					info.Country = countryName
				}
			}
			if record.City.Names != nil {
				if cityName, exists := record.City.Names["en"]; exists {
					info.City = cityName
				}
			}
		}
	}

	// Get ASN information
	p.asnMutex.RLock()
	asnDB := p.asnDB
	p.asnMutex.RUnlock()

	if asnDB != nil {
		if record, err := asnDB.ASN(ip); err == nil {
			if record.AutonomousSystemOrganization != "" {
				info.ASN = record.AutonomousSystemOrganization
			}
		}
	}

	return info
}

// ParseIPAddress is a convenience function that uses the singleton parser
func ParseIPAddress(ipStr string) IPInfo {
	parser := GetIPParser()
	return parser.ParseIP(ipStr)
}

// isPrivateIP checks if an IP address is private/local
func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Check for private IPv4 ranges
	if ip.To4() != nil {
		// 10.0.0.0/8
		if ip[0] == 10 {
			return true
		}
		// 172.16.0.0/12
		if ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31 {
			return true
		}
		// 192.168.0.0/16
		if ip[0] == 192 && ip[1] == 168 {
			return true
		}
	}

	// Check for private IPv6 ranges (fc00::/7)
	if len(ip) == 16 && (ip[0]&0xfe) == 0xfc {
		return true
	}

	return false
}

// Close closes the GeoIP databases
func (p *IPParser) Close() {
	p.cityMutex.Lock()
	if p.cityDB != nil {
		p.cityDB.Close()
		p.cityDB = nil
	}
	p.cityMutex.Unlock()

	p.asnMutex.Lock()
	if p.asnDB != nil {
		p.asnDB.Close()
		p.asnDB = nil
	}
	p.asnMutex.Unlock()
}