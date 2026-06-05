package main

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// startScheduler launches the background audit scheduler.
// It runs due schedules every minute and re-runs auto-detection daily.
func startScheduler(db *sql.DB) {
	go runScheduler(db)
}

func runScheduler(db *sql.DB) {
	_ = dbAutoDetectSchedules(db)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	dailyTicker := time.NewTicker(24 * time.Hour)
	defer dailyTicker.Stop()

	for {
		select {
		case now := <-ticker.C:
			runDueSchedules(db, now)
		case <-dailyTicker.C:
			_ = dbAutoDetectSchedules(db)
		}
	}
}

func runDueSchedules(db *sql.DB, now time.Time) {
	schedules, err := dbListDueSchedules(db, now)
	if err != nil {
		return
	}
	for _, s := range schedules {
		// Honour time window (UTC hours, -1 means any time)
		if s.TimeWindowStart >= 0 && s.TimeWindowEnd > s.TimeWindowStart {
			h := now.UTC().Hour()
			if h < s.TimeWindowStart || h >= s.TimeWindowEnd {
				continue
			}
		}

		auditID := uuid.New().String()
		p := auditParams{
			Provider: s.Provider,
			Model:    s.Model,
		}

		if err := dbCreateAudit(db, auditID, s.Repo, s.Model, s.Provider); err != nil {
			continue
		}

		nextRun := now.Add(time.Duration(s.IntervalH) * time.Hour)
		if err := dbUpdateScheduleLastRun(db, s.ID, now, nextRun); err != nil {
			continue
		}

		go runAudit(db, auditID, "", p)
	}
}
