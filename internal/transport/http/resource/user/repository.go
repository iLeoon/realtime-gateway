package user

import (
	"context"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/jackc/pgx/v5/pgxpool"
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
	const path errors.PathName = "user/repository"
	const op errors.Op = "repository.GetUserById"
	var user User

	err := r.db.QueryRow(ctx, `SELECT user_id, username, email FROM users WHERE user_id =$1`, userId).Scan(&user.UserId,
		&user.UserName,
		&user.Email)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}

	return &user, nil
}
