package application

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	httpadapter "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

// ExecutionStrategy encapsulates the runtime execution behavior for a specific Mode.
type ExecutionStrategy interface {
	Name() string
	Execute(app *App) error
}

// APIStrategy executes the application as an AWS Lambda function handling
// incoming HTTP traffic from API Gateway (payload format v2).
type APIStrategy struct{}

func (s *APIStrategy) Name() string {
	return "API Gateway Lambda"
}

func (s *APIStrategy) Execute(app *App) error {
	log.Println("Starting Lambda API handler (API Gateway v2 HTTP)...")
	adapter := httpadapter.NewV2(app.routes())
	lambda.Start(adapter.ProxyWithContext)
	return nil
}

// WorkerStrategy executes the application as a scheduled AWS Lambda worker
// triggered by Amazon EventBridge to refresh tournaments and dispatch notifications.
type WorkerStrategy struct{}

func (s *WorkerStrategy) Name() string {
	return "EventBridge Worker Lambda"
}

func (s *WorkerStrategy) Execute(app *App) error {
	log.Println("Starting Lambda Worker handler (EventBridge)...")
	lambda.Start(app.RunWorker)
	return nil
}

// ServerStrategy executes the application as a standalone long-running HTTP server
// with in-process cron scheduling and notification worker (for local development).
type ServerStrategy struct{}

func (s *ServerStrategy) Name() string {
	return "Standalone HTTP Server"
}

func (s *ServerStrategy) Execute(app *App) error {
	log.Printf("Starting standalone server on port: %s\n", app.Config.Port)
	app.TournamentJob.InitRefresh()
	go app.NotificationJob.Init()
	return http.ListenAndServe(":"+app.Config.Port, app.routes())
}

// NewStrategy creates an ExecutionStrategy based on the provided Mode enum.
func NewStrategy(mode Mode) (ExecutionStrategy, error) {
	switch mode {
	case ModeAPI:
		return &APIStrategy{}, nil
	case ModeWorker:
		return &WorkerStrategy{}, nil
	case ModeServer:
		return &ServerStrategy{}, nil
	default:
		return nil, fmt.Errorf("no execution strategy found for mode: %s", mode)
	}
}
