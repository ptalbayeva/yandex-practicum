package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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

	_, err := r.db.ExecContext(ctx, "INSERT INTO shorten_urls (uuid, original, shorten, user_id)"+
		"VALUES ($1, $2, $3, $4)", uuid.NewString(), u.Original, u.Code, u.UserID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = ErrConflict
		}
	}

	return err
}

func (r *DBRepository) FindByCode(code string) (*model.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	url := r.db.QueryRowContext(ctx,
		"SELECT original, shorten FROM shorten_urls  WHERE shorten = $1", code)

	var URL model.URL

	err := url.Scan(&URL.Original, &URL.Code)

	if err != nil {
		return nil, err
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

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO shorten_urls (uuid, original, shorten, user_id) VALUES ($1, $2, $3, $4)"+
		"ON CONFLICT (original) DO NOTHING")
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, u := range urls {
		_, fail := stmt.ExecContext(ctx, u.UID, u.Original, u.Code, u.UserID)
		if fail != nil {
			return fail
		}

	}

	return tx.Commit()
}

func (r *DBRepository) FindManyByUserId(userID string) ([]*model.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	rows, err := r.db.QueryContext(ctx,
		"SELECT uuid, original, shorten, user_id FROM shorten_urls WHERE user_id = $1", userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var urls []*model.URL

	for rows.Next() {
		var u model.URL

		err = rows.Scan(&u.UID, &u.Original, &u.Code, &u.UserID)
		if err != nil {
			return nil, err
		}

		urls = append(urls, &u)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return urls, nil
}
