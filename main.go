package main

import (
	"io"
	"log/slog"
	"os"
)


func setupLogger(env string, filename string) (*slog.Logger, error) {
    var log *slog.Logger
	var out io.Writer

	if filename != "" {
		file , err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}

		out = file
	} else {
		out = os.Stdout
	}

    switch env {
    case "local":
        log = slog.New(
			slog.NewTextHandler(out, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}),
		)
    case "production":
        log = slog.New(
			slog.NewJSONHandler(out, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
    }

    return log, nil
}

func main() {

}
