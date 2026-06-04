package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// ── Schedule handlers ─────────────────────────────────────────────────────────

type createScheduleRequest struct {
	Repo            string `json:"repo"`
	IntervalH       int    `json:"interval_h"`
	Provider        string `json:"provider"`
	Model           string `json:"model"`
	TimeWindowStart int    `json:"time_window_start"` // UTC hour 0-23, -1 = any
	TimeWindowEnd   int    `json:"time_window_end"`
	CliFallback     bool   `json:"cli_fallback"`
}

type updateScheduleRequest struct {
	IntervalH       int    `json:"interval_h"`
	Provider        string `json:"provider"`
	Model           string `json:"model"`
	Enabled         bool   `json:"enabled"`
	TimeWindowStart int    `json:"time_window_start"`
	TimeWindowEnd   int    `json:"time_window_end"`
}

func handleListSchedules(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schedules, err := dbListSchedules(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if schedules == nil {
			schedules = []scheduleRow{}
		}
		writeJSON(w, http.StatusOK, schedules)
	}
}

func handleCreateSchedule(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createScheduleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Repo == "" {
			http.Error(w, "invalid request body — repo is required", http.StatusBadRequest)
			return
		}
		if req.IntervalH <= 0 {
			req.IntervalH = defaultScheduleIntervalH
		}
		if req.TimeWindowStart == 0 && req.TimeWindowEnd == 0 {
			req.TimeWindowStart = -1
			req.TimeWindowEnd = -1
		}
		id := uuid.New().String()
		if err := dbCreateSchedule(db, id, req.Repo, req.IntervalH, req.Provider, req.Model,
			req.TimeWindowStart, req.TimeWindowEnd, req.CliFallback, false, ""); err != nil {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		schedules, err := dbListSchedules(db)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		for _, s := range schedules {
			if s.ID == id {
				writeJSON(w, http.StatusCreated, s)
				return
			}
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func handleUpdateSchedule(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var req updateScheduleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.IntervalH <= 0 {
			req.IntervalH = defaultScheduleIntervalH
		}
		if err := dbUpdateSchedule(db, id, req.IntervalH, req.Provider, req.Model,
			req.Enabled, req.TimeWindowStart, req.TimeWindowEnd); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleDeleteSchedule(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := dbDeleteSchedule(db, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleTriggerSchedule(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := dbTriggerScheduleNow(db, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
