package user

import (
	"context"
	"strconv"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/jackc/pgx/v5/pgxpool"
)

const path errors.PathName = "user/repository"

type repository struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetUserByID(userID string, ctx context.Context) (*User, error) {
	const op errors.Op = "repository.GetUserById"
	var user User

	// Linted and Optimized Query
	query := `
	    SELECT 
		u.user_id, 
		u.username, 
		u.email, 
		COALESCE(u.avatar_url, '') 
	    FROM users u 
	    WHERE u.user_id = $1
	    LIMIT 1
	`

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.UserID,
		&user.UserName,
		&user.Email,
		&user.Image,
	)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}

	return &user, nil
}

func (r *repository) GetFriends(ctx context.Context, userID string) (FriendsList, error) {
	const op errors.Op = "repository.GetFriends"
	var fl FriendsList

	rows, err := r.db.Query(ctx, `
		SELECT u.user_id, u.username, u.email, COALESCE(u.avatar_url, '')
		FROM friends f
		INNER JOIN users u on u.user_id = (
			CASE
				WHEN f.sender_id = $1 THEN f.recipient_id
				ELSE f.sender_id
			END
		)
		WHERE (f.sender_id = $1 OR f.recipient_id = $1)
		AND f.status = 'accepted'`, userID)
	if err != nil {
		return fl, apierror.DatabaseErrorClassification(path, op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.UserID, &u.UserName, &u.Email, &u.Image); err != nil {
			return fl, apierror.DatabaseErrorClassification(path, op, err)
		}
		fl.Value = append(fl.Value, u)
	}

	if err := rows.Err(); err != nil {
		return fl, apierror.DatabaseErrorClassification(path, op, err)
	}

	return fl, nil
}

func (r *repository) DeleteFriend(ctx context.Context, authenticatedID string, targetID string) error {
	const op errors.Op = "repository.DeleteFriend"

	authIDInt, err := strconv.Atoi(authenticatedID)
	if err != nil {
		return errors.B(path, op, errors.Client, "userID from auth context is not a valid integer")
	}

	targetIDInt, err := strconv.Atoi(targetID)
	if err != nil {
		return errors.B(path, op, errors.Client, "targetID path param is not a valid integer")
	}

	senderID := min(authIDInt, targetIDInt)
	recipientID := max(authIDInt, targetIDInt)

	tag, err := r.db.Exec(ctx,
		`DELETE FROM friends
		 WHERE sender_id = $1 AND recipient_id = $2 AND status = 'accepted'`,
		senderID, recipientID)
	if err != nil {
		return apierror.DatabaseErrorClassification(path, op, err)
	}

	if tag.RowsAffected() == 0 {
		return errors.B(path, op, errors.NotFound, "no accepted friendship found between these users")
	}

	return nil
}
