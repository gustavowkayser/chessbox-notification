package application

import (
	"fmt"
	"os"
	"strings"
)

// Mode represents the runtime execution mode of the application.
type Mode string

const (
	// ModeAPI handles HTTP requests via AWS API Gateway and Lambda.
	ModeAPI Mode = "api"

	// ModeWorker runs periodic scheduled scraping and push notification dispatching via EventBridge.
	ModeWorker Mode = "worker"

	// ModeServer runs a standalone long-running HTTP server with local cron for development.
	ModeServer Mode = "server"
)

// String returns the string representation of the mode.
func (m Mode) String() string {
	return string(m)
}

// IsValid reports whether the mode is one of the recognized modes.
func (m Mode) IsValid() bool {
	switch m {
	case ModeAPI, ModeWorker, ModeServer:
		return true
	default:
		return false
	}
}

// ParseMode parses a string into an application Mode enum.
func ParseMode(val string) (Mode, error) {
	normalized := Mode(strings.ToLower(strings.TrimSpace(val)))
	if normalized.IsValid() {
		return normalized, nil
	}
	return "", fmt.Errorf("unknown application mode: %q (expected '%s', '%s', or '%s')",
		val, ModeAPI, ModeWorker, ModeServer)
}

// DetectMode determines the appropriate runtime Mode from explicit configuration
// or infers it from environment indicators (such as AWS Lambda function name).
func DetectMode(configured string) (Mode, error) {
	if configured == "" {
		lambdaFunc := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
		if lambdaFunc != "" {
			return ParseMode(strings.Split(lambdaFunc, "-")[strings.Count(lambdaFunc, "-")])
		}
		return ModeServer, nil
	}

	mode, err := ParseMode(configured)
	return mode, err
}
