package notification

import (
	"context"
	"log"
)

type Job struct {
	service        *Service
	notificationCh chan *Notification
}

func NewJob(service *Service, notificationCh chan *Notification) *Job {
	return &Job{
		service:        service,
		notificationCh: notificationCh,
	}
}

func (j *Job) Init() {
	for job := range j.notificationCh {
		err := j.service.SendNotification(
			context.Background(), 
			job.TournamentID,
		)

		if err != nil {
			log.Printf("Error occurred when pushing notification: %s\n", err.Error())
		}
	}
}

func (j *Job) Send(ctx context.Context, tournamentID string) error {
	return j.service.SendNotification(ctx, tournamentID)
}
