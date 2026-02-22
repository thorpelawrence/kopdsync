package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/thorpelawrence/kopdsync/internal/database"
	"github.com/thorpelawrence/kopdsync/internal/opds"
	"github.com/thorpelawrence/kopdsync/internal/sync"
)

var (
	listen            = flag.String("listen", ":8080", "address and port to listen on (e.g., ':8080', '127.0.0.1:8080')")
	dsn               = flag.String("db", "sync.db", "sqlite database file for sync")
	booksDir          = flag.String("books", "./books", "directory containing EPUB files for OPDS")
	openRegistrations = flag.Bool("registrations", false, "allow new user registrations")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	db, err := database.OpenDB(*dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	mux := http.NewServeMux()

	opds.RegisterRoutes(mux, db, &opds.Config{
		BooksDir: *booksDir,
	})

	sync.RegisterRoutes(mux, db, &sync.Config{
		OpenRegistrations: *openRegistrations,
	})

	http.Handle("/", mux)

	slog.Info("starting", "listen", *listen)

	if err := http.ListenAndServe(*listen, nil); err != nil {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}
