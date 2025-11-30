package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/yandex-practicum/shorten-url/internal/model"
)

type DBRepository struct {
	db *sql.DB
}

func NewDBRepository(db *sql.DB) *DBRepository {
	return &DBRepository{db: db}
}

func (r *DBRepository) Save(u *model.URL) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, "INSERT INTO shorten_urls (uuid, original, shorten)"+
		"VALUES ($1, $2, $3)", uuid.NewString(), u.Original, u.Code)
	if err != nil {
		return err
	}

	return nil
}

func (r *DBRepository) FindByCode(code string) (*model.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	url := r.db.QueryRowContext(ctx,
		"SELECT original, shorten FROM shorten_urls  WHERE shorten = $1", code)

	var URL model.URL

	err := url.Scan(&URL.Original, &URL.Code)

	if err != nil {
		log.Println(url)
		log.Println(err)
		return nil, errors.New("error while scanning url")
	}

	return &URL, nil
}
