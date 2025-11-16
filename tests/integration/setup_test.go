package integration

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"google.golang.org/grpc/codes"

	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/app"
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/database"
)

var (
	testUserDBConf = &database.Config{
		Driver: "postgres",
		Addr: os.Getenv("PG_ADDR"),
		User: os.Getenv("PG_USER"),
		Password: os.Getenv("PG_PASSWORD"),
		DBName: os.Getenv("TEST_DB_NAME"),
	}

	testSessionDBConf = &database.Config{
		Driver: "redis",
		Addr: os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DBName: os.Getenv("TEST_REDIS_NUM"),
	}
)

type TestCase struct {
	name string
	args any
	code codes.Code
	wantErr bool
}

func TestMain(m *testing.M) {
	var err error

	if err = godotenv.Load(); err != nil {
		log.Fatalf("Failed to find .env file: %v", err)
	}

	makeMigrations(testUserDBConf)

	if err = runTestApp(4445); err != nil {
		log.Fatalf("Can't run test app - %v", err)
	}

	code := m.Run()

	os.Exit(code)
}

func makeMigrations(pgConfig *database.Config) {
	m, err := database.NewMigrator(pgConfig, os.Getenv("MIGRATIONS_DIR"))
	if err != nil {
		log.Fatalf("migrator creation error - %v\n", err)
	}

	if err = m.Apply(); err != nil {
		log.Fatalf("migrations apply failure - %v\n", err)
	}

	log.Println("INFO: migrations applied successfully!")
}

func runTestApp(port int) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	testApp, err := app.New(
		logger,
		testUserDBConf,
		testSessionDBConf,
		port,
	)

	fmt.Println("TEST DATABASE HAS BEEN CREATED!!!")

	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	go func() {
		if err = testApp.Run(); err != nil {
			log.Fatalf("Error during test app running: %v", err)
		}
	}()
	
	return nil
}
