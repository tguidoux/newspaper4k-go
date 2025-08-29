package helpers

import (
	"testing"
	"time"
)

func TestCreateHTTPClient(t *testing.T) {
	tests := []struct {
		name            string
		timeoutSeconds  int
		expectedTimeout time.Duration
	}{
		{
			name:            "positive timeout",
			timeoutSeconds:  5,
			expectedTimeout: 5 * time.Second,
		},
		{
			name:            "zero timeout",
			timeoutSeconds:  0,
			expectedTimeout: 0,
		},
		{
			name:            "large timeout",
			timeoutSeconds:  300,
			expectedTimeout: 300 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := CreateHTTPClient(tt.timeoutSeconds)
			if client == nil {
				t.Fatal("Expected non-nil client")
			}
			if client.Timeout != tt.expectedTimeout {
				t.Errorf("Expected timeout %v, got %v", tt.expectedTimeout, client.Timeout)
			}
		})
	}
}

func TestCreateDefaultHTTPClient(t *testing.T) {
	client := CreateDefaultHTTPClient()
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	expectedTimeout := time.Duration(DefaultTimeoutSeconds) * time.Second
	if client.Timeout != expectedTimeout {
		t.Errorf("Expected timeout %v, got %v", expectedTimeout, client.Timeout)
	}
}
