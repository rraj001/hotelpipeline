package application

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type CronJobService struct {
	cron     *cron.Cron
	fetcher  *HotelFetcher
	entryID  cron.EntryID
	interval string
}

func NewCronJobService(fetcher *HotelFetcher, interval string) *CronJobService {
	return &CronJobService{
		cron:     cron.New(cron.WithSeconds()),
		fetcher:  fetcher,
		interval: interval,
	}
}

func (cs *CronJobService) Start() error {
	entryID, err := cs.cron.AddFunc(cs.interval, func() {
		if err := cs.fetchJob(); err != nil {
			log.Printf("Cron job failed: %v", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	cs.entryID = entryID
	cs.cron.Start()

	return nil
}

func (cs *CronJobService) Stop() {
	if cs.cron != nil {
		cs.cron.Stop()
	}
}

func (cs *CronJobService) RunOnce() error {
	return cs.fetchJob()
}

func (cs *CronJobService) fetchJob() error {
	if err := cs.fetcher.FetchAndProcess(); err != nil {
		log.Printf("Scheduled hotel fetch failed: %v", err)
		return err
	}

	return nil
}

func (cs *CronJobService) GetNextRun() time.Time {
	if cs.cron == nil {
		return time.Time{}
	}

	entry := cs.cron.Entry(cs.entryID)
	return entry.Next
}

func (cs *CronJobService) GetStatus() map[string]interface{} {
	if cs.cron == nil {
		return map[string]interface{}{
			"running":  false,
			"next_run": nil,
		}
	}

	entry := cs.cron.Entry(cs.entryID)
	return map[string]interface{}{
		"running":  !entry.Next.IsZero(),
		"next_run": entry.Next,
		"interval": cs.interval,
	}
}
