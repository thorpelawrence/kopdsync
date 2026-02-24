package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func Init(debug bool) {
	w := os.Stderr
	opts := &tint.Options{
		NoColor:    !isatty.IsTerminal(w.Fd()),
		TimeFormat: time.TimeOnly,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			err, ok := a.Value.Any().(error)
			if !ok {
				return a
			}

			key := a.Key
			a = tint.Err(err)
			a.Key = key
			return a
		},
	}
	if debug {
		opts.Level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, opts)))
}
