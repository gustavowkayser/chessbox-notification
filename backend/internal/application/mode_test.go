package application

import (
	"os"
	"testing"
)

func TestParseMode(t *testing.T) {
	tests := []struct {
		input       string
		expected    Mode
		expectError bool
	}{
		{"api", ModeAPI, false},
		{"API", ModeAPI, false},
		{"worker", ModeWorker, false},
		{"WORKER", ModeWorker, false},
		{"server", ModeServer, false},
		{" Server ", ModeServer, false},
		{"unknown", "", true},
		{"", "", true},
	}

	for _, tc := range tests {
		result, err := ParseMode(tc.input)
		if tc.expectError {
			if err == nil {
				t.Errorf("ParseMode(%q): expected error, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseMode(%q): unexpected error: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("ParseMode(%q): expected %s, got %s", tc.input, tc.expected, result)
			}
		}
	}
}

func TestDetectMode(t *testing.T) {
	// Test explicit configuration
	mode, err := DetectMode("worker")
	if err != nil || mode != ModeWorker {
		t.Fatalf("expected ModeWorker, got %v (err: %v)", mode, err)
	}

	// Test fallback to server when no lambda env var is set
	os.Unsetenv("AWS_LAMBDA_FUNCTION_NAME")
	mode, err = DetectMode("")
	if err != nil || mode != ModeServer {
		t.Fatalf("expected ModeServer, got %v (err: %v)", mode, err)
	}

	// Test inferring api from Lambda env
	os.Setenv("AWS_LAMBDA_FUNCTION_NAME", "chessbox-notification-production-api")
	mode, err = DetectMode("")
	if err != nil || mode != ModeAPI {
		t.Fatalf("expected ModeAPI, got %v (err: %v)", mode, err)
	}

	// Test inferring worker from Lambda env
	os.Setenv("AWS_LAMBDA_FUNCTION_NAME", "chessbox-notification-production-worker")
	mode, err = DetectMode("")
	if err != nil || mode != ModeWorker {
		t.Fatalf("expected ModeWorker, got %v (err: %v)", mode, err)
	}

	os.Unsetenv("AWS_LAMBDA_FUNCTION_NAME")
}

func TestNewStrategy(t *testing.T) {
	tests := []struct {
		mode         Mode
		expectedName string
		expectError  bool
	}{
		{ModeAPI, "API Gateway Lambda", false},
		{ModeWorker, "EventBridge Worker Lambda", false},
		{ModeServer, "Standalone HTTP Server", false},
		{Mode("invalid"), "", true},
	}

	for _, tc := range tests {
		strategy, err := NewStrategy(tc.mode)
		if tc.expectError {
			if err == nil {
				t.Errorf("NewStrategy(%q): expected error, got nil", tc.mode)
			}
		} else {
			if err != nil {
				t.Errorf("NewStrategy(%q): unexpected error: %v", tc.mode, err)
			}
			if strategy.Name() != tc.expectedName {
				t.Errorf("NewStrategy(%q): expected name %q, got %q", tc.mode, tc.expectedName, strategy.Name())
			}
		}
	}
}
