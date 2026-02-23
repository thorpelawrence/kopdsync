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

	mux.HandleFunc("GET /users/auth", s.WithAuth(s.Auth))
	mux.HandleFunc("POST /users/create", s.CreateUser)

	mux.HandleFunc("GET /syncs/progress/{document}", s.WithAuth(s.GetProgress))
	mux.HandleFunc("PUT /syncs/progress", s.WithAuth(s.UpdateProgress))
}
