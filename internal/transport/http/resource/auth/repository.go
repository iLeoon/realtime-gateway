package auth

import (
	"context"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/pkg/log"
	"github.com/jackc/pgx/v5"
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

// CreateOrUpdateUser checks if the user already exists in the db, if the user exists update the information as needed.
// If not create a new record for the user, and return userId.
func (r *repository) CreateOrUpdateUser(ctx context.Context, pi ProviderIdentity) (user *User, err error) {
	user = &User{}
	const path errors.PathName = "auth/repository"
	const op errors.Op = "repository.CreateOrUpdateUser"

	var tx pgx.Tx
	tx, err = r.db.Begin(ctx)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)

	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	err = tx.QueryRow(ctx, `INSERT INTO users (username, email)
			VALUES ($1, $2) ON CONFLICT (email) DO UPDATE SET username = EXCLUDED.username RETURNING user_id, email, username`,
		pi.Name,
		pi.Email).Scan(&user.UserID, &user.Email, &user.UserName)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}

	_, err = tx.Exec(ctx, `INSERT INTO providers (provider, provider_user_id, user_id)
			VALUES ($1, $2, $3) ON CONFLICT (provider, provider_user_id) DO NOTHING`,
		pi.Provider,
		pi.ProviderID,
		user.UserID,
	)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)

	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}

	log.Info.Printf("user upserted: id=%s email=%s", user.UserID, user.Email)
	return user, nil
}
