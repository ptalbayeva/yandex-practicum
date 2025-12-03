package repository

import (
	"context"
	"database/sql"
	"errors"
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
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := r.db.ExecContext(ctx, "INSERT INTO shorten_urls (uuid, original, shorten)"+
		"VALUES ($1, $2, $3)", uuid.NewString(), u.Original, u.Code)
	if err != nil {
		return err
	}

	return nil
}

func (r *DBRepository) FindByCode(code string) (*model.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	url := r.db.QueryRowContext(ctx,
		"SELECT original, shorten FROM shorten_urls  WHERE shorten = $1", code)

	var URL model.URL

	err := url.Scan(&URL.Original, &URL.Code)

	if err != nil {
		return nil, errors.New("error while scanning url")
	}

	return &URL, nil
}

func (r *DBRepository) Close() error {
	return r.db.Close()
}

func (r *DBRepository) SaveMany(urls []*model.URL) error {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	tx, err := r.db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO shorten_urls (uuid, original, shorten) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, u := range urls {
		_, fail := stmt.ExecContext(ctx, u.UID, u.Original, u.Code)
		if fail != nil {
			return fail
		}
	}

	return tx.Commit()
}
