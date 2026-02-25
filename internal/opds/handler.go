package opds

import (
	"database/sql"
	"net/http"
)

type Server struct {
	db  *sql.DB
	cfg *Config
}

func RegisterRoutes(mux *http.ServeMux, db *sql.DB, cfg *Config) {
	s := Server{db: db, cfg: cfg}

	mux.Handle("GET /catalog", s.WithBasicAuth(http.HandlerFunc(s.Catalog)))
	mux.Handle("GET /files/", s.WithBasicAuth(
		http.StripPrefix("/files/", http.FileServer(http.Dir(s.cfg.BooksDir))),
	))
}
