package sync

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Document struct {
	Device     string  `json:"device"`
	DeviceID   string  `json:"device_id"`
	Document   string  `json:"document"`
	Percentage float64 `json:"percentage"`
	Progress   string  `json:"progress"`
	Timestamp  int64   `json:"timestamp"`
}

func (s *Server) GetProgress(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-Auth-User")

	docID := r.PathValue("document")
	if docID == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var doc Document
	row := s.db.QueryRow(`
		SELECT
			device,
			device_id,
			document,
			percentage,
			progress,
			timestamp
		FROM
			progress
		WHERE
			document = ?
			AND username = ?
	`, docID, username)
	if err := row.Scan(
		&doc.Device,
		&doc.DeviceID,
		&doc.Document,
		&doc.Percentage,
		&doc.Progress,
		&doc.Timestamp,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		slog.Error("retrieving status", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(doc); err != nil {
		slog.Error("writing response json", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

type UpdateProgressResponse struct {
	Document  string `json:"document"`
	Timestamp int64  `json:"timestamp"`
}

func (s *Server) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-Auth-User")

	var doc Document
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		slog.Error("reading request json", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if doc.Document == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if doc.Timestamp == 0 {
		doc.Timestamp = time.Now().Unix()
	}

	if _, err := s.db.Exec(`
		INSERT INTO progress (
			device,
			device_id,
			document,
			percentage,
			progress,
			timestamp,
			username
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (document, username) DO UPDATE
		SET
			device = EXCLUDED.device,
			device_id = EXCLUDED.device_id,
			percentage = EXCLUDED.percentage,
			progress = EXCLUDED.progress,
			timestamp = EXCLUDED.timestamp
	`,
		doc.Device,
		doc.DeviceID,
		doc.Document,
		doc.Percentage,
		doc.Progress,
		doc.Timestamp,
		username,
	); err != nil {
		slog.Error("upserting progress", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(UpdateProgressResponse{
		Document:  doc.Document,
		Timestamp: doc.Timestamp,
	}); err != nil {
		slog.Error("writing response json", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
