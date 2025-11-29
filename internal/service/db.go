package service

import (
	"context"
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type DBService struct {
	dsn string
}

func NewDBService(dsn string) *DBService {
	return &DBService{dsn}
}

func (s *DBService) Ping() error {
	db, err := sql.Open("sqlite", s.dsn)
	if err != nil {
		return err
	}

	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		return err
	}

	return nil
}
