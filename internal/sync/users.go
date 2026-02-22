package sync

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.OpenRegistrations {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	var user CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		slog.Error("decoding request json", "error", err)
		return
	}
	defer r.Body.Close()

	if _, err := s.db.Exec(
		`INSERT INTO users (username, password) VALUES (?, ?)`,
		user.Username,
		user.Password,
	); err != nil {
		if sqlErr, ok := errors.AsType[*sqlite.Error](err); ok {
			if sqlErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
				http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
				return
			}
		}
		slog.Error("creating user in database", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
