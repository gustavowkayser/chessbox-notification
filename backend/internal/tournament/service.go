package tournament

import (
	"chessbox-notification/internal/notification"
	"context"
	"errors"
	"log"
	"strings"
	"sync"
)

type Service struct {
	provider   Provider
	repository Repository
}

func NewService(provider Provider, repository Repository) *Service {
	return &Service{
		provider:   provider,
		repository: repository,
	}
}

func (s *Service) Refresh(
	ctx context.Context,
	notificationCh chan *notification.Notification,
) {
	// Runs every minute
	log.Println("Refresh tournaments...")

	jobs := make(chan *Tournament)

	var wg sync.WaitGroup

	for _ = range 10 {
		wg.Add(1)

		wg.Go(func() {
			defer wg.Done()
			s.Worker(ctx, jobs, notificationCh)
		})
	}

	tournaments, err := s.repository.GetAllActive(ctx)

	if err != nil {
		log.Printf("Error occurred on refresh job: %s\n", err.Error())
		close(jobs)
		wg.Wait()
		return
	}

	log.Printf("Found %d tournaments\n", len(*tournaments))

	for i := range *tournaments {
		select {
		case jobs <- &(*tournaments)[i]:
		case <-ctx.Done():
			break
		}
	}

	close(jobs)

	wg.Wait()
}

func (s *Service) Worker(
	ctx context.Context,
	jobs chan *Tournament,
	notificationCh chan *notification.Notification,
) {
	for {
		select {
		case tournament, ok := <-jobs:
			if !ok {
				return
			}

			log.Printf("Processing tournament: %s\n", tournament.ID)

			updated, err := s.provider.GetTournament(tournament.URL)

			if err != nil {
				continue
			}

			if tournament.CurrentRound == updated.CurrentRound {
				continue
			}

			err = s.repository.Update(ctx, tournament.ID, updated)

			notificationCh <- &notification.Notification{
				TournamentID: tournament.ID,
			}

			if err != nil {
				log.Printf("Error occured when updating tournament: %s\n", err.Error())
			}

		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) SearchTournaments(ctx context.Context, query string) (*[]TournamentView, error) {
	query = strings.TrimSpace(query)

	if query == "" {
		return nil, errors.New("Search query must not be empty")
	}

	tournaments, err := s.provider.SearchTournaments(query)

	if err != nil {
		return nil, err
	}

	return tournaments, nil
}
