package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	UserNotFoundErr = errors.New("No user found in the database with that userId")
	UnexpectedErr   = errors.New("An unexpected error occured while processing the request.")
)

type repository struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetUserById(userId string, ctx context.Context) (*User, error) {
	var user User
	err := r.db.QueryRow(ctx, `SELECT user_id, username, email FROM users WHERE user_id =$1`, userId).Scan(&user.UserId, &user.UserName, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w:%d", UserNotFoundErr, err)
		} else {
			return nil, fmt.Errorf("%w:%d", UnexpectedErr, err)
		}
	}
	return &user, nil
}
