package friendrequest

import (
	"context"
	"strconv"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const path errors.PathName = "friendrequest/repository"

type repository struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, userID string, body FriendRequestBody) (*FriendRequest, error) {
	const op errors.Op = "repository.Create"

	userIDToInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, errors.B(path, op, errors.Client, "userID from auth context is not a valid integer")
	}

	// sender_id is always the lower ID — prevents duplicate composite-key entries.
	senderID := min(userIDToInt, body.RecipientID)
	recipientID := max(userIDToInt, body.RecipientID)

	fr := &FriendRequest{}
	err = r.db.QueryRow(ctx,
		`INSERT INTO friends (sender_id, recipient_id, author_id, status)
		 VALUES ($1, $2, $3, 'pending')
		 RETURNING sender_id, recipient_id, author_id, status, created_at`,
		senderID, recipientID, userIDToInt,
	).Scan(&fr.SenderID, &fr.RecipientID, &fr.AuthorID, &fr.Status, &fr.CreatedAt)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}

	return fr, nil
}

// friendRequestListQuery is the shared JOIN used by both list queries.
// The caller supplies the WHERE clause via the filter argument.
const friendRequestListQuery = `
	SELECT
		s.user_id,  s.username,  s.email,  COALESCE(s.avatar_url, ''),
		rc.user_id, rc.username, rc.email, COALESCE(rc.avatar_url, ''),
		a.user_id,  a.username,  a.email,  COALESCE(a.avatar_url, ''),
		f.status,
		f.created_at
	FROM friends f
	JOIN users s  ON s.user_id  = f.sender_id
	JOIN users rc ON rc.user_id = f.recipient_id
	JOIN users a  ON a.user_id  = f.author_id
`

func scanFriendRequestList(rows pgx.Rows, path errors.PathName, op errors.Op) (FriendRequestList, error) {
	var fl FriendRequestList
	defer rows.Close()

	for rows.Next() {
		var item FriendRequestDetailed
		if err := rows.Scan(
			&item.Sender.ID, &item.Sender.DisplayName, &item.Sender.Email, &item.Sender.Image,
			&item.Recipient.ID, &item.Recipient.DisplayName, &item.Recipient.Email, &item.Recipient.Image,
			&item.Creator.ID, &item.Creator.DisplayName, &item.Creator.Email, &item.Creator.Image,
			&item.Status,
			&item.CreatedAt,
		); err != nil {
			return fl, apierror.DatabaseErrorClassification(path, op, err)
		}
		fl.Value = append(fl.Value, item)
	}

	if err := rows.Err(); err != nil {
		return fl, apierror.DatabaseErrorClassification(path, op, err)
	}
	return fl, nil
}

func (r *repository) FindSentRequests(ctx context.Context, userID string) (FriendRequestList, error) {
	const op errors.Op = "repository.FindSentRequests"

	rows, err := r.db.Query(ctx,
		friendRequestListQuery+`WHERE f.author_id = $1 AND status = 'pending'`,
		userID)
	if err != nil {
		return FriendRequestList{}, apierror.DatabaseErrorClassification(path, op, err)
	}

	return scanFriendRequestList(rows, path, op)
}

func (r *repository) FindReceivedRequests(ctx context.Context, userID string) (FriendRequestList, error) {
	const op errors.Op = "repository.FindReceivedRequests"

	rows, err := r.db.Query(ctx,
		friendRequestListQuery+`
		WHERE f.author_id != $1
		  AND (f.sender_id = $1 OR f.recipient_id = $1) AND status = 'pending'`,
		userID)
	if err != nil {
		return FriendRequestList{}, apierror.DatabaseErrorClassification(path, op, err)
	}

	return scanFriendRequestList(rows, path, op)
}
func (r *repository) AcceptRequest(ctx context.Context, userID string, authorID string) error {
	const op errors.Op = "repository.AcceptRequest"

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return errors.B(path, op, errors.Client, "userID from auth context is not a valid integer")
	}
	authorIDInt, err := strconv.Atoi(authorID)
	if err != nil {
		return errors.B(path, op, errors.Client, "authorID path param is not a valid integer")
	}

	senderID := min(userIDInt, authorIDInt)
	recipientID := max(userIDInt, authorIDInt)

	tag, err := r.db.Exec(ctx,
		`UPDATE friends SET status = 'accepted'
		 WHERE sender_id = $1 AND recipient_id = $2 AND author_id = $3 AND status = 'pending'`,
		senderID, recipientID, authorIDInt)

	if err != nil {
		return apierror.DatabaseErrorClassification(path, op, err)
	}

	if tag.RowsAffected() == 0 {
		return errors.B(path, op, errors.NotFound, "no pending request found to accept")
	}
	return nil
}

func (r *repository) CancelSentRequest(ctx context.Context, authorID string, targetID string) error {
	const op errors.Op = "repository.CancelSentRequest"

	authorIDInt, err := strconv.Atoi(authorID)
	if err != nil {
		return errors.B(path, op, errors.Client, "userID from auth context is not a valid integer")
	}
	targetIDInt, err := strconv.Atoi(targetID)
	if err != nil {
		return errors.B(path, op, errors.Client, "targetID path param is not a valid integer")
	}

	senderID := min(authorIDInt, targetIDInt)
	recipientID := max(authorIDInt, targetIDInt)

	tag, err := r.db.Exec(ctx,
		`DELETE FROM friends
		 WHERE sender_id = $1 AND recipient_id = $2 AND author_id = $3`,
		senderID, recipientID, authorIDInt)
	if err != nil {
		return apierror.DatabaseErrorClassification(path, op, err)
	}
	if tag.RowsAffected() == 0 {
		return errors.B(path, op, errors.NotFound, "no pending sent request found to cancel")
	}
	return nil
}

func (r *repository) DeclineReceivedRequest(ctx context.Context, userID string, targetID string) error {
	const op errors.Op = "repository.DeclineReceivedRequest"

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return errors.B(path, op, errors.Client, "userID from auth context is not a valid integer")
	}
	targetIDInt, err := strconv.Atoi(targetID)
	if err != nil {
		return errors.B(path, op, errors.Client, "targetID path param is not a valid integer")
	}

	senderID := min(userIDInt, targetIDInt)
	recipientID := max(userIDInt, targetIDInt)

	tag, err := r.db.Exec(ctx,
		`DELETE FROM friends
		 WHERE sender_id = $1 AND recipient_id = $2 AND author_id != $3`,
		senderID, recipientID, userIDInt)
	if err != nil {
		return apierror.DatabaseErrorClassification(path, op, err)
	}
	if tag.RowsAffected() == 0 {
		return errors.B(path, op, errors.NotFound, "no pending received request found to decline")
	}
	return nil
}
