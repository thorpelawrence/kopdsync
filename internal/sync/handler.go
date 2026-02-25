package sync

import (
	"database/sql"
	"net/http"
)

type Server struct {
	db  *sql.DB
	cfg *Config
}

func RegisterRoutes(mux *http.ServeMux, db *sql.DB, cfg *Config) {
	s := &Server{
		db:  db,
		cfg: cfg,
	}

	mux.Handle("GET /users/auth", s.WithAuth(http.HandlerFunc(s.Auth)))
	mux.HandleFunc("POST /users/create", s.CreateUser)

	mux.Handle("GET /syncs/progress/{document}", s.WithAuth(http.HandlerFunc(s.GetProgress)))
	mux.Handle("PUT /syncs/progress", s.WithAuth(http.HandlerFunc(s.UpdateProgress)))
}
