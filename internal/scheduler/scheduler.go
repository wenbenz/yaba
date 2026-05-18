package scheduler

import (
	"context"
	"log"
	"time"
)

// Start runs job.Run once at startup and then once every 24 hours. It blocks
// until ctx is cancelled.
func Start(ctx context.Context, job *ReminderJob) {
	run := func() {
		if err := job.Run(ctx); err != nil {
			log.Printf("scheduler: reminder job error: %v", err)
		}
	}

	run()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			run()
		case <-ctx.Done():
			return
		}
	}
}
