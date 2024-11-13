package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func ConnectClickHouse() (*sql.DB, error) {
	conn, err := sql.Open("clickhouse", "http://localhost:8123/stock_ticker?username=default&password=")

	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Set connection options (optional)
	conn.SetConnMaxIdleTime(5 * time.Minute)
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)

	// Verify connection by pinging the database
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return conn, nil
}
