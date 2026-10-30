package application

import (
	"chessbox-notification/internal/database"
	"chessbox-notification/internal/device"
	"chessbox-notification/internal/middleware"
	"chessbox-notification/internal/notification"
	"chessbox-notification/internal/subscription"
	"chessbox-notification/internal/tournament"
	"context"
	"database/sql"
	"log"
	"net/http"
	"sync"
)

type App struct {
	Config Config
	DB     *sql.DB

	DeviceHandler       *device.Handler
	SubscriptionHandler *subscription.Handler
	TournamentHandler   *tournament.Handler
	NotificationHandler *notification.Handler
	AuthMiddleware      func(http.Handler) http.Handler
	LogMiddleware       func(http.Handler) http.Handler
	TournamentJob       *tournament.Job
	NotificationJob     *notification.Job
}

func New() (*App, error) {
	config := LoadConfig()

	db, err := database.NewPostgres(config.DatabaseURL)
	if err != nil {
		return nil, err
	}

	err = database.Migrate(db, context.Background())
	if err != nil {
		return nil, err
	}

	argon2config := device.ArgonConfig{
		Salt:       []byte(config.Secret),
		TimeCost:   2,
		MemoryCost: 64 * 1024,
		Threads:    4,
		KeyLength:  32,
	}

	chessProvider := tournament.NewProvider()
	notificationProvider := notification.NewExponent()

	notificationRepository := notification.NewRepository(db)
	notificationService := notification.NewService(notificationRepository, notificationProvider)
	notificationHandler := notification.NewHandler(notificationService)

	deviceRepository := device.NewRepository(db)
	deviceService := device.NewService(deviceRepository, argon2config)
	deviceHandler := device.NewHandler(deviceService)

	tournamentRepository := tournament.NewRepository(db)

	subscriptionRepository := subscription.NewRepository(db)
	subscriptionService := subscription.NewService(subscriptionRepository, tournamentRepository, chessProvider)
	subscriptionHandler := subscription.NewHandler(subscriptionService)

	authMiddleware := middleware.DeviceAuth(deviceService)
	logMiddleware := middleware.LogMiddleware()

	notificationCh := make(chan *notification.Notification)

	tournamentService := tournament.NewService(chessProvider, tournamentRepository)
	tournamentJob := tournament.NewJob(tournamentService, notificationCh)
	notificationJob := notification.NewJob(notificationService, notificationCh)
	tournamentHandler := tournament.NewHandler(tournamentService)

	return &App{
		Config:              config,
		DB:                  db,
		DeviceHandler:       deviceHandler,
		SubscriptionHandler: subscriptionHandler,
		NotificationHandler: notificationHandler,
		AuthMiddleware:      authMiddleware,
		LogMiddleware:       logMiddleware,
		TournamentJob:       tournamentJob,
		NotificationJob:     notificationJob,
		TournamentHandler:   tournamentHandler,
	}, nil
}

// Run executes the application using the strategy matching Config.AppMode.
func (app *App) Run() error {
	strategy, err := NewStrategy(app.Config.AppMode)
	if err != nil {
		return err
	}

	log.Printf("Executing application with %s strategy (mode: %s)\n", strategy.Name(), app.Config.AppMode)
	return strategy.Execute(app)
}

// RunWorker executes a single pass of tournament refresh and push notifications
// suitable for AWS Lambda triggered by EventBridge.
func (app *App) RunWorker(ctx context.Context) error {
	log.Println("Worker execution started...")

	notificationCh := make(chan *notification.Notification, 50)
	var notifWg sync.WaitGroup

	// Consume notifications concurrently during this worker execution
	notifWg.Add(1)
	go func() {
		defer notifWg.Done()
		for notif := range notificationCh {
			err := app.NotificationJob.Send(ctx, notif.TournamentID)
			if err != nil {
				log.Printf("Error sending notification for tournament %s: %s\n", notif.TournamentID, err.Error())
			}
		}
	}()

	// Scrape and update tournaments
	app.TournamentJob.RefreshWithChannel(ctx, notificationCh)

	// Signal notification consumer and wait for all notifications to finish sending
	close(notificationCh)
	notifWg.Wait()

	log.Println("Worker execution completed successfully.")
	return nil
}
