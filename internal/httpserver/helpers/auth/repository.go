package auth

import (
	"context"
	"errors"

	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateOrUpdateUser(ctx context.Context, user ProviderIdentity) (userId int, err error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}

// CreateOrUpdateUser checks if the user already exists in the db, if the user exists update the information as needed.
// If not create a new record for the user, and return userId.
func (r *repository) CreateOrUpdateUser(ctx context.Context, user ProviderIdentity) (userId int, err error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Retrive the user from database who is trying to login.
	err = tx.QueryRow(ctx, `select user_id
	from providers 
	where provider_user_id =$1 and provider =$2`, user.ProviderID, user.Provider).Scan(&userId)

	if err != nil {
		// If there is no rows retrieve from the database,
		// Create records in the users, and providers tables.
		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(ctx, `INSERT INTO users (username, email)
		VALUES ($1, $2) RETURNING user_id`,
				user.Name,
				user.Email).Scan(&userId)

			if err != nil {
				return 0, err
			}

			_, err = tx.Exec(ctx, `INSERT INTO providers (provider, provider_user_id, user_id)
		VALUES ($1, $2, $3)`,
				user.Provider,
				user.ProviderID,
				userId,
			)
			if err != nil {
				return 0, err
			}

			logger.Info(`Created a new user with their corresponding provider into the database`)
		} else {
			return 0, err
		}

	} else {
		// If the user who is attempting to login already exists
		// Update their data only if the username updated otherwise skip.
		var tag pgconn.CommandTag
		tag, err = tx.Exec(ctx, `UPDATE users SET username=$1
	         WHERE user_id=$2 AND (username <> $1)`,
			user.Name,
			userId,
		)

		if err != nil {
			return 0, err
		}

		if tag.RowsAffected() != 0 {
			logger.Info("A user has updated their data")
		}

	}
	err = tx.Commit(ctx)
	return userId, err
}
