package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/iLeoon/realtime-gateway/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepositpryInterface interface {
	HandleLogins(context.Context, models.ProviderUser) error
}

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

func (ar *AuthRepository) HandleLogins(ctx context.Context, User models.ProviderUser) error {
	tx, beginErr := ar.db.Begin(ctx)
	if beginErr != nil {
		return beginErr
	}

	var userID int
	defer tx.Rollback(ctx)

	// Retrive the user from database
	// who is trying to login.
	execErr := tx.QueryRow(ctx, `select user_id
	from providers 
	where provider_user_id =$1 and provider =$2`, User.ProviderID, User.Provider).Scan(&userID)

	// If there is no user in the database
	// no row is retrived, then create the user.
	if errors.Is(execErr, pgx.ErrNoRows) {
		createUserErr := tx.QueryRow(ctx, `INSERT INTO users (username, email)
		VALUES ($1, $2) RETURNING user_id`,
			User.Name,
			User.Email).Scan(&userID)

		if createUserErr != nil {
			return createUserErr
		}

		_, createProviderErr := tx.Exec(ctx, `INSERT INTO providers (provider, provider_user_id, user_id)
		VALUES ($1, $2, $3)`,
			User.Provider,
			User.ProviderID,
			userID,
		)
		if createProviderErr != nil {
			return createProviderErr
		}

		logger.Info(`Created a new user with their corresponding provider into the database`)
	}

	// Update the user data only if the username or email updated otherwise skip.
	row, updateUserErr := tx.Exec(ctx, `UPDATE users SET email=$1, username=$2 
	WHERE user_id=$3 AND (username <> $2 OR email <> $1)`,
		User.Email,
		User.Name,
		userID,
	)

	if updateUserErr != nil {
		return updateUserErr
	}

	fmt.Println(row.RowsAffected())

	return tx.Commit(ctx)
}
