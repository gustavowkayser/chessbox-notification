package tournament

import (
	"chessbox-notification/internal/notification"
	"context"

	"github.com/robfig/cron/v3"
)

type Job struct {
	service        *Service
	notificationCh chan *notification.Notification
	cron           *cron.Cron
}

func NewJob(service *Service, notificationCh chan *notification.Notification) *Job {
	return &Job{
		service:        service,
		notificationCh: notificationCh,
	}
}

func (j *Job) InitRefresh() {
	j.cron = cron.New()
	j.cron.AddFunc("@every 1m", j.Refresh)
	j.cron.Start()
}

func (j *Job) Stop() {
	if j.cron != nil {
		j.cron.Stop()
	}
}

func (j *Job) Refresh() {
	ctx := context.Background()
	j.service.Refresh(ctx, j.notificationCh)
}

func (j *Job) RefreshWithChannel(ctx context.Context, ch chan *notification.Notification) {
	j.service.Refresh(ctx, ch)
}
