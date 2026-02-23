package sync

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string
	Password string
}

func (s *Server) WithAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Header.Get("X-Auth-User")
		password := r.Header.Get("X-Auth-Key")

		if username == "" || password == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, MessageUnauthorized)
			return
		}

		var user User
		row := s.db.QueryRow(`SELECT username, password FROM users WHERE username = ?`, username)
		err := row.Scan(&user.Username, &user.Password)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			slog.Error("checking for existing user", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		userExists := !errors.Is(err, sql.ErrNoRows)
		if !userExists {
			if !s.cfg.OpenRegistrations {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintln(w, MessageForbidden)
				return
			}

			// auto create user when registrations are open
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				slog.Error("generating password hash", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if _, err := s.db.Exec(`INSERT INTO users (username, password) VALUES (?, ?)`, username, hash); err != nil {
				slog.Error("creating user", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		} else {
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintln(w, MessageUnauthorized)
				return
			}
		}

		h(w, r)
	}
}
