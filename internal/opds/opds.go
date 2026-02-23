package opds

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	epub "github.com/taylorskalyo/goreader/epub"
)

type AtomFeed struct {
	XMLName   xml.Name    `xml:"feed"`
	Xmlns     string      `xml:"xmlns,attr"`
	XmlnsDc   string      `xml:"xmlns:dc,attr"`
	XmlnsOpds string      `xml:"xmlns:opds,attr"`
	ID        string      `xml:"id"`
	Title     string      `xml:"title"`
	Updated   string      `xml:"updated"`
	Author    *AtomAuthor `xml:"author"`
	Link      []AtomLink  `xml:"link"`
	Entry     []AtomEntry `xml:"entry"`
}

type AtomAuthor struct {
	Name string `xml:"name"`
	URI  string `xml:"uri,omitempty"`
}

type AtomLink struct {
	Rel   string `xml:"rel,attr"`
	Href  string `xml:"href,attr"`
	Type  string `xml:"type,attr"`
	Title string `xml:"title,attr,omitempty"`
}

type AtomEntry struct {
	ID       string         `xml:"id"`
	Title    string         `xml:"title"`
	Author   *AtomAuthor    `xml:"author"`
	Updated  string         `xml:"updated"`
	Issued   string         `xml:"dc:issued"`
	Link     []AtomLink     `xml:"link"`
	Category []AtomCategory `xml:"category"`
	Summary  string         `xml:"summary,omitempty"`
}

type AtomCategory struct {
	Scheme string `xml:"scheme,attr"`
	Label  string `xml:"label,attr,omitempty"`
}

type EPUBMetadata struct {
	Author          string
	Description     string
	PublicationDate string
	Subject         string
	Title           string
}

func NewEPUBMetadata(file io.ReaderAt, info fs.FileInfo) (md *EPUBMetadata, err error) {
	rc, err := epub.NewReader(file, info.Size())
	if err != nil {
		return nil, fmt.Errorf("creating epub reader: %w", err)
	}

	if len(rc.Rootfiles) == 0 {
		return nil, fmt.Errorf("no rootfiles found in epub")
	}

	metadata := rc.Rootfiles[0].Metadata

	publicationDate := info.ModTime().Format(time.RFC3339)
	if len(metadata.Event) > 0 {
		publicationDate = metadata.Event[0].Date
	}

	return &EPUBMetadata{
		Author:          metadata.Creator,
		Description:     metadata.Description,
		PublicationDate: publicationDate,
		Subject:         metadata.Subject,
		Title:           metadata.Title,
	}, nil
}

func (s *Server) Catalog(w http.ResponseWriter, r *http.Request) {
	scheme := "http"
	if r.URL.Scheme != "" {
		scheme = r.URL.Scheme
	} else if r.Header.Get("X-Forwarded-Proto") != "" {
		scheme = r.Header.Get("X-Forwarded-Proto")
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

	feed := AtomFeed{
		Xmlns:     "http://www.w3.org/2005/Atom",
		XmlnsDc:   "http://purl.org/dc/terms/",
		XmlnsOpds: "http://opds-spec.org/2010/catalog",
		ID:        fmt.Sprintf("urn:feed:%s", baseURL),
		Title:     "kopdsync catalog",
		Updated:   time.Now().Format(time.RFC3339),
		Author: &AtomAuthor{
			Name: "kopdsync",
			URI:  baseURL,
		},
		Link: []AtomLink{
			{
				Rel:  "start",
				Href: "/opds/catalog",
				Type: "application/atom+xml;profile=opds-catalog;kind=navigation",
			},
			{
				Rel:  "self",
				Href: "/opds/catalog",
				Type: "application/atom+xml;profile=opds-catalog;kind=navigation",
			},
		},
	}

	err := filepath.WalkDir(s.cfg.BooksDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("accessing path during content scan", "path", path, "error", err)
			return nil
		}

		if strings.HasPrefix(d.Name(), ".") { // skip hidden files
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(d.Name()), ".epub") {
			return nil
		}

		feedEntry, err := getFeedEntry(d, path, s)
		if err != nil {
			return fmt.Errorf("getting feed entry '%v': %w", path, err)
		}

		feed.Entry = append(feed.Entry, *feedEntry)

		return nil
	})
	if err != nil {
		path := s.cfg.BooksDir
		if pErr, ok := errors.AsType[*pathError](err); ok {
			path = pErr.Path()
		}
		slog.Error("scanning books directory", "path", path, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	baseDir := filepath.Base(s.cfg.BooksDir)
	feed.Author.Name = baseDir
	feed.Title = baseDir

	w.Header().Set("Content-Type", "application/atom+xml;profile=opds-catalog;charset=utf-8")
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		slog.Error("writing xml header", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := xml.NewEncoder(w).Encode(feed); err != nil {
		slog.Error("encode opds feed xml", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

type pathError struct {
	Err  string
	path string
}

func (f *pathError) Error() string {
	return f.Err
}

func (f *pathError) Path() string {
	return f.path
}

func newPathError(path string, format string, a ...any) error {
	return &pathError{
		Err:  fmt.Sprintf(format, a...),
		path: path,
	}
}

func getFeedEntry(d fs.DirEntry, path string, s *Server) (*AtomEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, newPathError(path, "opening file: %s", err)
	}
	defer f.Close()

	info, err := d.Info()
	if err != nil {
		return nil, newPathError(path, "getting file info: %s", err)
	}

	md, err := NewEPUBMetadata(f, info)
	if err != nil {
		return nil, newPathError(path, "getting epub metadata: %s", err)
	}

	if md.Title == "" {
		md.Title = strings.TrimSuffix(d.Name(), filepath.Ext(path))
	}

	relPath, err := filepath.Rel(s.cfg.BooksDir, path)
	if err != nil {
		return nil, newPathError(path, "getting relative path: %s", err)
	}

	entry := AtomEntry{
		ID:    fmt.Sprintf("urn:file:%s", relPath),
		Title: md.Title,
		Author: &AtomAuthor{
			Name: md.Author,
		},
		Updated: info.ModTime().Format(time.RFC3339),
		Issued:  md.PublicationDate,
		Link: []AtomLink{
			{
				Rel:   "http://opds-spec.org/acquisition",
				Href:  fmt.Sprintf("/files/%s", relPath),
				Type:  "application/epub+zip",
				Title: md.Title,
			},
		},
		Category: []AtomCategory{
			{
				Scheme: "http://purl.org/ontology/library/subject",
				Label:  md.Subject,
			},
		},
		Summary: md.Description,
	}

	return &entry, nil
}
