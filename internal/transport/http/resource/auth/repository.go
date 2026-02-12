package auth

import (
	"context"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
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

	err = r.db.QueryRow(ctx, `SELECT users.user_id, email, username FROM providers INNER JOIN users ON users.user_id = providers.user_id
	WHERE provider_user_id =$1 AND provider =$2`, pi.ProviderID, pi.Provider).Scan(&user.UserID, &user.Email, &user.UserName)
	if err == nil {
		// If the user who is attempting to login already exists
		// Update their data only if the username updated otherwise skip.
		_, err = r.db.Exec(ctx, `UPDATE users SET username=$1
	         WHERE user_id=$2 AND (username <> $1)`,
			pi.Name,
			user.UserID,
		)

		if err != nil {
			return nil, apierror.DatabaseErrorClassification(path, op, err)

		}
		return user, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}

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
			VALUES ($1, $2) RETURNING user_id, email, username`,
		pi.Name,
		pi.Email).Scan(&user.UserID, &user.Email, &user.UserName)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}

	_, err = tx.Exec(ctx, `INSERT INTO providers (provider, provider_user_id, user_id)
			VALUES ($1, $2, $3)`,
		pi.Provider,
		pi.ProviderID,
		user.UserID,
	)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)

	}
	logger.Info(`Created a new user with their corresponding provider into the database`)

	err = tx.Commit(ctx)

	return user, apierror.DatabaseErrorClassification(path, op, err)
}
