package opds

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/thorpelawrence/kopdsync/internal/logger"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string
	Password string
}

func md5Hex(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func (s *Server) WithBasicAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logger.FromContext(r.Context())

		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="kopdsync"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		var user User
		row := s.db.QueryRow(`
			SELECT username, password
			FROM users
			WHERE username = ?
		`, username)
		err := row.Scan(&user.Username, &user.Password)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.Header().Set("WWW-Authenticate", `Basic realm="kopdsync"`)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			logger.Error("checking for existing user", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// sync clients hash passwords with MD5, reproduce that when comparing
		md5Password := md5Hex(password)

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(md5Password)); err != nil {
			w.Header().Set("WWW-Authenticate", `Basic realm="kopdsync"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}
