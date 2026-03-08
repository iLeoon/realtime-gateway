package conversation

import (
	"context"
	"encoding/json"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
	"github.com/jackc/pgx/v5/pgxpool"
)

const path errors.PathName = "conversation/repository"

type repository struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateConversation(ctx context.Context, creatorId string, cr ConversationRequest) (*ConversationCreatedResponse, error) {
	const op errors.Op = "repository.CreateConversation"

	c := &ConversationCreatedResponse{}
	var isExist bool

	if cr.ConversationType == "private-chat" {
		err := r.db.QueryRow(ctx, `
	SELECT EXISTS (
	    SELECT 1
	    FROM users_conversations uc1
	    JOIN users_conversations uc2 ON uc1.conversation_id = uc2.conversation_id
	    JOIN conversations c ON c.conversation_id = uc1.conversation_id
	    WHERE uc1.user_id = $1
	      AND uc2.user_id = $2
	      AND c.conversation_type = 'private-chat'
	)`, creatorId, cr.ParticipantIDs[0]).Scan(&isExist)
		if err != nil {
			return nil, apierror.DatabaseErrorClassification(path, op, err)
		}

		if isExist {
			return nil, errors.B(path, op, errors.Client, "private conversation already exists between these users")
		}
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	// Create a conversation with it's creator.
	err = tx.QueryRow(ctx,
		`INSERT INTO conversations (creator_id, conversation_type, group_name) VALUES($1, $2, $3)
		RETURNING conversation_id, creator_id, conversation_type, group_name, created_at`,
		creatorId, cr.ConversationType, cr.GroupName).Scan(&c.ConversationId, &c.CreatorID, &c.ConversationType, &c.GroupName, &c.CreatedAt)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}
	// Insert the creator of the conversation
	_, err = tx.Exec(ctx, `INSERT INTO users_conversations (conversation_id, user_id) VALUES($1, $2)`, c.ConversationId, creatorId)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}
	// Insert all members wether the chat was private or group
	for _, v := range cr.ParticipantIDs {
		// Insert both creator and recipient into users_conversations.
		_, err = tx.Exec(ctx, `INSERT INTO users_conversations (conversation_id, user_id) VALUES($1, $2)`, c.ConversationId, v)
		if err != nil {
			return nil, apierror.DatabaseErrorClassification(path, op, err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}
	return c, nil
}

func (r *repository) FindConversation(ctx context.Context, conversationId string, userId string) (*Conversation, error) {
	const op errors.Op = "repository.FindConversation"
	c := &Conversation{}

	err := r.db.QueryRow(ctx, `
	SELECT 
		c.conversation_id,
		c.creator_id,
		c.conversation_type,
		c.group_name,
		c.created_at
        FROM conversations c
        JOIN users_conversations uc ON uc.conversation_id = c.conversation_id AND uc.user_id = $2
        WHERE c.conversation_id = $1`, conversationId, userId).Scan(
		&c.ConversationId,
		&c.CreatorID,
		&c.ConversationType,
		&c.GroupName,
		&c.CreatedAt,
	)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}

	p, err := r.FetchMembers(ctx, conversationId)
	if err != nil {
		return nil, errors.B(path, op, err)
	}

	c.Participants = p
	return c, nil
}

func (r *repository) FindConversations(ctx context.Context, userId string) (ConversationsList, error) {
	var conversations ConversationsList
	const op errors.Op = "repository.FindConversations"

	rows, err := r.db.Query(ctx, `
	SELECT 
	    c.conversation_id, 
	    c.creator_id, 
	    c.conversation_type, 
	    c.group_name,
	    c.created_at,
	    (
		SELECT json_agg(
		    json_build_object(
			'id',		u.user_id::TEXT,
			'displayName',  u.username,
			'email',	u.email,
			'displayImage', COALESCE(u.avatar_url, ''),
			'joinedDate',	(uc2.joined_at AT TIME ZONE 'UTC'),
			'role',		CASE WHEN c.creator_id = uc2.user_id THEN 'owner' ELSE 'member' END
		    )
		)
		FROM users_conversations uc2
		JOIN users u ON u.user_id = uc2.user_id
		WHERE uc2.conversation_id = c.conversation_id
	    ) AS participants
	FROM conversations c
	WHERE EXISTS (
	    SELECT 1
	    FROM users_conversations uc
	    WHERE uc.conversation_id = c.conversation_id
	      AND uc.user_id = $1
	)`, userId)
	if err != nil {
		return conversations, apierror.DatabaseErrorClassification(path, op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var c Conversation
		var participantsRaw []byte
		err := rows.Scan(
			&c.ConversationId,
			&c.CreatorID,
			&c.ConversationType,
			&c.GroupName,
			&c.CreatedAt,
			&participantsRaw,
		)

		if err != nil {
			return conversations, apierror.DatabaseErrorClassification(path, op, err)
		}

		if err := json.Unmarshal(participantsRaw, &c.Participants); err != nil {
			return conversations, errors.B(path, op, "failed to unmarshal participants", errors.Internal)
		}
		conversations.Value = append(conversations.Value, c)
	}

	if err := rows.Err(); err != nil {
		return conversations, apierror.DatabaseErrorClassification(path, op, err)
	}

	return conversations, nil
}

func (r *repository) FindMembers(ctx context.Context, conversationId string, userId string) (ParticipantsList, error) {
	const op errors.Op = "repository.FindMembers"
	var pl ParticipantsList
	var isExist bool

	// Check if the user is authorized to see the conversation participants or not.
	err := r.db.QueryRow(ctx, `
	    SELECT EXISTS (
	        SELECT 1 FROM users_conversations
	        WHERE conversation_id = $1 AND user_id = $2
	    )`, conversationId, userId).Scan(&isExist)
	if err != nil {
		return pl, apierror.DatabaseErrorClassification(path, op, err)
	}

	// EXISTS returns a boolean value
	// true if there was a row or false if there was not
	// that's why we have to explicitly specifiy the error if there was no record.
	if !isExist {
		return pl, errors.B(path, op, errors.NotFound, "user is not a participant of this conversation")
	}
	// Fetch the conversation participants.
	p, err := r.FetchMembers(ctx, conversationId)
	if err != nil {
		return pl, errors.B(path, op, err)
	}
	pl.Value = p
	return pl, nil
}

func (r *repository) UpdateParticipants(ctx context.Context, conversationId string, requesterId string, body UpdateConversationRequest) (ParticipantsList, error) {
	const op errors.Op = "repository.UpdateParticipants"
	var pl ParticipantsList

	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM users_conversations
			WHERE conversation_id = $1 AND user_id = $2
		)`, conversationId, requesterId).Scan(&exists)
	if err != nil {
		return pl, apierror.DatabaseErrorClassification(path, op, err)
	}
	if !exists {
		return pl, errors.B(path, op, errors.NotFound, "user is not a participant of this conversation")
	}

	var conversationType string
	err = r.db.QueryRow(ctx,
		`SELECT conversation_type FROM conversations WHERE conversation_id = $1`,
		conversationId).Scan(&conversationType)
	if err != nil {
		return pl, apierror.DatabaseErrorClassification(path, op, err)
	}
	if conversationType != "group-chat" {
		return pl, errors.B(path, op, errors.Client, "cannot add participants to a private-chat")
	}

	for _, id := range body.ParticipantIDs {
		_, err = r.db.Exec(ctx,
			`INSERT INTO users_conversations (conversation_id, user_id) VALUES($1, $2) ON CONFLICT DO NOTHING`,
			conversationId, id)
		if err != nil {
			return pl, apierror.DatabaseErrorClassification(path, op, err)
		}
	}

	participants, err := r.FetchMembers(ctx, conversationId)
	if err != nil {
		return pl, errors.B(path, op, err)
	}
	pl.Value = participants
	return pl, nil
}

// FetchMembers is a helper function that fetches all the participants
// of a conversation including the conversation creator
// and assign "owner" for the conversation creator
// and "member" for the rest of the conversation members
func (r *repository) FetchMembers(ctx context.Context, conversationId string) ([]Participant, error) {
	const op errors.Op = "repository.FetchMembers"
	var ps []Participant

	rows, err := r.db.Query(ctx, `
	    SELECT 
		u.user_id, 
		u.username, 
		u.email,
		COALESCE(u.avatar_url, ''),
		uc.joined_at,
		CASE WHEN c.creator_id = uc.user_id THEN 'owner' ELSE 'member' END AS role
	    FROM users_conversations uc
	    JOIN users u ON u.user_id = uc.user_id
	    JOIN conversations c ON c.conversation_id = uc.conversation_id
	    WHERE uc.conversation_id = $1
	`, conversationId)
	if err != nil {
		return ps, apierror.DatabaseErrorClassification(path, op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var p Participant
		err := rows.Scan(
			&p.UserId,
			&p.UserName,
			&p.Email,
			&p.Image,
			&p.JoinedAt,
			&p.Role,
		)
		if err != nil {
			return ps, apierror.DatabaseErrorClassification(path, op, err)
		}
		ps = append(ps, p)
	}

	if err := rows.Err(); err != nil {
		return ps, apierror.DatabaseErrorClassification(path, op, err)
	}
	return ps, nil
}
