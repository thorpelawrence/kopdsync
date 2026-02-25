package opds

import (
	"net/http"
)

func (s *Server) FileServer(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/files/", http.FileServer(http.Dir(s.cfg.BooksDir))).ServeHTTP(w, r)
}
