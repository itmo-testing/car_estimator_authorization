package integration

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/domain"
	"golang.org/x/crypto/bcrypt"
)

func TestPGConn(t *testing.T) *sql.DB {
	t.Helper()

    db, err := sql.Open(
		testUserDBConf.Driver, 
		testUserDBConf.GetPgConnString(false),
	)
	
    if err != nil {
        t.Fatalf("Failed to connect to test DB: %v", err)
    }
    
    if err := db.Ping(); err != nil {
        t.Fatalf("Failed to ping test DB: %v", err)
    }
    
    return db
}

func TestRedisConn(t *testing.T) *redis.Client {
	t.Helper()

	options, err := redis.ParseURL(testSessionDBConf.GetRedisConnString())
	if err != nil {
		t.Fatalf("Error while parsing redis url: %v", err)
	}

	client := redis.NewClient(options)
	ctx := context.Background()

	if _, err := client.Ping(ctx).Result(); err != nil {
		t.Fatalf("Test redis connection error: %v", err)
	}

	return client
}

func CleanUpTestStorages(t *testing.T, db *sql.DB, rd *redis.Client) {
	t.Helper()

	if db != nil {
		// Очищаем postgres
		_, err := db.Exec("DELETE FROM users;")
		if err != nil {
			t.Logf("Warning: failed to cleanup test pg data: %v", err)
		}
		
		if err := db.Close(); err != nil {
			t.Logf("Warning: failed to close pg connection: %v", err)
		}
	}
	

	if rd != nil {
		// очищаем redis
		ctx := context.Background()
		err := rd.FlushDB(ctx).Err()

		if err != nil {
			t.Logf("Warning: failed to cleanup test redis data: %v", err)
		}

		if err = rd.Close(); err != nil {
			t.Logf("Warning: failed to close redis connection: %v", err)
		}
	}

	t.Logf("Finished storages cleanup")
}

func CreateTestUser(t *testing.T, db *sql.DB, user *domain.User) {
	t.Helper()

	var (
		id = uuid.New()
		now = time.Now()
		hash, _ = bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		query = "INSERT INTO users (id, fullName, email, phone, password, birthDate, registerDate) VALUES ($1, $2, $3, $4, $5, $6, $7);"
	)
	
	_, err := db.Exec(query, id, user.FullName, user.Email, user.Phone, hash, user.BirthDate, now)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
}
