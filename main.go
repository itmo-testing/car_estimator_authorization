package main

import (
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/app"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/database"
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
	if err := godotenv.Load(); err != nil {
		log.Fatalln("error: can't find .env file")
	}

	logger, err := setupLogger(os.Getenv("MODE"), "")
	if err != nil {
		log.Fatalln("error: can't setup logger")
	}

	pgConfig := &database.Config{
		Driver: "postgres",
		Addr: os.Getenv("PG_ADDR"),
		User: os.Getenv("PG_USER"),
		Password: os.Getenv("PG_PASSWORD"),
		DBName: os.Getenv("PG_NAME"),
	}

	redisConfig := &database.Config{
		Driver: "redis",
		Addr: os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DBName: os.Getenv("REDIS_DB_NUM"),
	}

	m, err := database.NewMigrator(pgConfig, os.Getenv("MIGRATIONS_DIR"))
	if err != nil {
		log.Fatalf("migrator creation error - %v\n", err)
	}

	if err = m.Apply(); err != nil {
		log.Fatalf("migrations apply failure - %v\n", err)
	}

	logger.Info("migrations applied successfully!")

	application, err := app.New(logger, pgConfig, redisConfig, 4444)
	if err != nil {
		log.Fatalf("application creation failed - %v\n", err)
	}

	err = application.Run()
	if err != nil {
		log.Fatalf("gRPC server shut down due to unexpected error - %v\n", err)
	}

	stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

    <-stop

    application.Stop()
    log.Print("Gracefully stopped")
}
