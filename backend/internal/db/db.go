package db

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Open(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetConnMaxLifetime(5 * time.Minute)
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	return conn, nil
}
