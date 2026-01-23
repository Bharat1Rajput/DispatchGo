package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)
	// the lazy opener (sql.Open does not establish any connections to the database.validate the arguments and returns a *DB)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// RunMigrations applies necessary database migrations
// func RunMigrations(db *sql.DB) error {
// 	schema := `CREATE TABLE IF NOT EXISTS tasks (
//     id UUID PRIMARY KEY,
//     type VARCHAR(50) NOT NULL,
//     payload JSONB NOT NULL,
//     status VARCHAR(20) NOT NULL,
//     retries INTEGER DEFAULT 0,
//     max_retries INTEGER DEFAULT 3,
//     error TEXT,
//     created_at TIMESTAMP NOT NULL,
//     updated_at TIMESTAMP NOT NULL
// );

// CREATE INDEX idx_tasks_status ON tasks(status);
// CREATE INDEX idx_tasks_created_at ON tasks(created_at DESC);

// `
// 	_, err := db.Exec(schema)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
