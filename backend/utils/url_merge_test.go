package utils

import (
	"testing"
)

func TestMergeURLWithShortURL(t *testing.T) {
	tests := []struct {
		name          string
		originalURL   string
		requestPath   string
		requestQuery  string
		expectedURL   string
		expectError   bool
	}{
		{
			name:         "Merge subpath and query with priority",
			originalURL:  "https://microsoft.com/docs/office/?source=ref&page=1",
			requestPath:  "/docs/faq",
			requestQuery: "source=apt",
			expectedURL:  "https://microsoft.com/docs/office/faq?page=1&source=apt",
			expectError:  false,
		},
		{
			name:         "Only subpath, no query",
			originalURL:  "https://example.com/api",
			requestPath:  "/shortcode/v1/users",
			requestQuery: "",
			expectedURL:  "https://example.com/api/v1/users",
			expectError:  false,
		},
		{
			name:         "Only query, no subpath",
			originalURL:  "https://example.com/page?existing=value",
			requestPath:  "/shortcode",
			requestQuery: "new=param",
			expectedURL:  "https://example.com/page?existing=value&new=param",
			expectError:  false,
		},
		{
			name:         "Query parameter override",
			originalURL:  "https://example.com/?color=red&size=large",
			requestPath:  "/shortcode",
			requestQuery: "color=blue",
			expectedURL:  "https://example.com/?color=blue&size=large",
			expectError:  false,
		},
		{
			name:         "No subpath or query",
			originalURL:  "https://example.com/page",
			requestPath:  "/shortcode",
			requestQuery: "",
			expectedURL:  "https://example.com/page",
			expectError:  false,
		},
		{
			name:         "Complex path merge",
			originalURL:  "https://example.com/docs/",
			requestPath:  "/shortcode/api/reference",
			requestQuery: "version=2",
			expectedURL:  "https://example.com/docs/api/reference?version=2",
			expectError:  false,
		},
		{
			name:         "URL without path",
			originalURL:  "https://example.com",
			requestPath:  "/shortcode/newpath",
			requestQuery: "param=value",
			expectedURL:  "https://example.com/newpath?param=value",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeURLWithShortURL(tt.originalURL, tt.requestPath, tt.requestQuery)
			
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if !tt.expectError && result != tt.expectedURL {
				t.Errorf("Expected URL: %s, got: %s", tt.expectedURL, result)
			}
		})
	}
}
