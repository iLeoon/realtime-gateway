package message

import (
	"context"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/resource/models"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/jackc/pgx/v5/pgxpool"
)

const path errors.PathName = "message/repository"

type repository struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *repository {
	return &repository{db: db}
}

func (r *repository) FindMessages(ctx context.Context, conversationID string, userID string) (models.MessagesList, error) {
	const op errors.Op = "repository.FindMessages"
	var ml models.MessagesList

	// Verify the user has access to this conversation, then fetch messages.
	rows, err := r.db.Query(ctx, `
	SELECT 
	 m.message_id, 
	 m.creator_id, 
	 m.conversation_id, 
	 m.content, 
	 m.created_at,
	 m.edited_at
	FROM messages m
	JOIN users_conversations uc ON m.conversation_id = uc.conversation_id
	WHERE m.conversation_id = $1 
	AND uc.user_id = $2      
	AND m.deleted_at IS NULL  
	ORDER BY m.created_at ASC
	`, conversationID, userID)
	if err != nil {
		return ml, apierror.DatabaseErrorClassification(path, op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.MessageID, &m.CreatorID, &m.ConversationID, &m.Content, &m.CreatedAt, &m.EditedAt); err != nil {
			return ml, apierror.DatabaseErrorClassification(path, op, err)
		}
		ml.Value = append(ml.Value, m)
	}

	if err := rows.Err(); err != nil {
		return ml, apierror.DatabaseErrorClassification(path, op, err)
	}

	return ml, nil
}
