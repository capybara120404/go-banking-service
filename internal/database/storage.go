package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Storage struct {
	DB *sql.DB
}

func NewStorage(dsn string, maxConnections int, minConnections int, maxConnIdleTime int) (*Storage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(minConnections)
	db.SetConnMaxIdleTime(time.Duration(maxConnIdleTime) * time.Second)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to the database")

	return &Storage{DB: db}, nil
}

func (s *Storage) Connection() *sql.DB {
	return s.DB
}

func (s *Storage) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}

	return nil
}
