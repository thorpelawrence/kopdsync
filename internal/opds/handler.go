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

	mux.HandleFunc("/catalog", s.WithBasicAuth(s.Catalog))
	mux.HandleFunc("/files/", s.WithBasicAuth(s.FileServer))
}
