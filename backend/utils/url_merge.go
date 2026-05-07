package utils

import (
	"net/url"
	"path"
	"strings"
)

// MergeURLWithShortURL merges the subpath and query parameters from the short URL
// into the original URL. Short URL parameters take priority over original URL parameters.
//
// Example:
//   Short URL: aka.ci/docs/faq?source=apt
//   Original URL: microsoft.com/docs/office/?source=ref&page=1
//   Result: microsoft.com/docs/office/faq?source=apt&page=1
func MergeURLWithShortURL(originalURL string, requestPath string, requestQuery string) (string, error) {
	// Parse the original URL
	parsedOriginal, err := url.Parse(originalURL)
	if err != nil {
		return "", err
	}

	// Extract the subpath from the request path (everything after the short code)
	// The requestPath should be something like "/shortcode/docs/faq"
	// We need to extract "/docs/faq" part
	var additionalPath string
	pathParts := strings.SplitN(requestPath, "/", 3)
	if len(pathParts) > 2 {
		additionalPath = "/" + pathParts[2]
	}

	// Merge the paths
	if additionalPath != "" {
		// Clean up the original path and join with the additional path
		originalPath := parsedOriginal.Path
		if originalPath == "" {
			originalPath = "/"
		}
		
		// Remove trailing slash from original path if present
		originalPath = strings.TrimSuffix(originalPath, "/")
		
		// Combine paths using path.Join for clean merging
		mergedPath := path.Join(originalPath, additionalPath)
		parsedOriginal.Path = mergedPath
	}

	// Merge query parameters
	if requestQuery != "" {
		// Parse the original URL's query parameters
		originalQuery := parsedOriginal.Query()
		
		// Parse the short URL's query parameters
		shortQuery, err := url.ParseQuery(requestQuery)
		if err != nil {
			return "", err
		}

		// Add original query parameters first
		mergedQuery := url.Values{}
		for key, values := range originalQuery {
			mergedQuery[key] = values
		}

		// Override with short URL parameters (short URL has priority)
		for key, values := range shortQuery {
			mergedQuery[key] = values
		}

		parsedOriginal.RawQuery = mergedQuery.Encode()
	}

	return parsedOriginal.String(), nil
}
