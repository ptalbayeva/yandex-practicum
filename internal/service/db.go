package service

import (
	"context"
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

func Ping(dsn string) error {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}

	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		return err
	}

	return nil
}
