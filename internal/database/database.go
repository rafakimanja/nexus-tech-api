package database

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Service represents the application's GORM database connection.
type Service interface {
	DB() *gorm.DB
	Health() map[string]string
	Close() error
}

type service struct {
	db *gorm.DB
}

var (
	dbInstance *service
	dbOnce     sync.Once
	dbErr      error
)

// New opens and reuses the application's GORM connection to PostgreSQL.
func New() Service {
	dbOnce.Do(func() {
		var dsn string
		localDB, err := strconv.ParseBool(os.Getenv("LOCAL_DB"))
		if err != nil && os.Getenv("LOCAL_DB") != "" {
			dbErr = fmt.Errorf("invalid LOCAL_DB value: %w", err)
			return
		}

		if localDB {
			connURL := url.URL{
				Scheme: "postgres",
				User:   url.UserPassword(os.Getenv("BLUEPRINT_DB_USERNAME"), os.Getenv("BLUEPRINT_DB_PASSWORD")),
				Host:   fmt.Sprintf("%s:%s", os.Getenv("BLUEPRINT_DB_HOST"), os.Getenv("BLUEPRINT_DB_PORT")),
				Path:   os.Getenv("BLUEPRINT_DB_DATABASE"),
			}
			query := connURL.Query()
			query.Set("sslmode", "disable")
			if schema := os.Getenv("BLUEPRINT_DB_SCHEMA"); schema != "" {
				query.Set("search_path", schema)
			}
			connURL.RawQuery = query.Encode()
			dsn = connURL.String()
		} else {
			dsn = os.Getenv("DATABASE_URL")
			if dsn == "" {
				dbErr = fmt.Errorf("DATABASE_URL must be set when LOCAL_DB is not enabled")
				return
			}
		}

		var db *gorm.DB
		db, dbErr = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if dbErr == nil {
			dbInstance = &service{db: db}
		}
	})

	if dbErr != nil {
		log.Fatal(dbErr)
	}
	return dbInstance
}

// DB returns the underlying GORM handle for queries and migrations.
func (s *service) DB() *gorm.DB {
	return s.db
}

// Health checks the database connection and reports connection-pool statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	stats := make(map[string]string)
	sqlDB, err := s.db.DB()
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		return stats
	}
	if err = sqlDB.PingContext(ctx); err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"
	poolStats := sqlDB.Stats()
	stats["open_connections"] = strconv.Itoa(poolStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(poolStats.InUse)
	stats["idle"] = strconv.Itoa(poolStats.Idle)
	stats["wait_count"] = strconv.FormatInt(poolStats.WaitCount, 10)
	stats["wait_duration"] = poolStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(poolStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(poolStats.MaxLifetimeClosed, 10)

	if poolStats.OpenConnections > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}
	if poolStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}
	if poolStats.MaxIdleClosed > int64(poolStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}
	if poolStats.MaxLifetimeClosed > int64(poolStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the underlying SQL connection pool.
func (s *service) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
