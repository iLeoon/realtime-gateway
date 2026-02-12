package conversation

import (
	"context"
	"encoding/json"
	"fmt"

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

func (r *repository) CreateConversation(ctx context.Context, creatorId string, cr ConversationRequest) (*Conversation, error) {
	const op errors.Op = "repository.CreateConversation"
	c := &Conversation{}
	var isExist bool

	err := r.db.QueryRow(ctx, `
	SELECT EXISTS (
	    SELECT 1
	    FROM conversations c
	    WHERE c.conversation_type = 'private-chat'
	      AND (
		    (
			c.creator_id = $1
			AND EXISTS (
			    SELECT 1
			    FROM users_conversations uc
			    WHERE uc.conversation_id = c.conversation_id
			      AND uc.user_id = $2
			)
		    )

		    OR

		    (
			c.creator_id = $2
			AND EXISTS (
			    SELECT 1
			    FROM users_conversations uc
			    WHERE uc.conversation_id = c.conversation_id
			      AND uc.user_id = $1
			)
		    )
	      )
	)`, creatorId, cr.RecipientId).Scan(&isExist)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}

	if isExist {
		return nil, errors.B(path, op, errors.Client)
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
		`
		INSERT INTO conversations (creator_id, conversation_type, last_message_id) VALUES($1, $2, $3)
		RETURNING conversation_id, creator_id, conversation_type, last_message_id`,
		creatorId, cr.ConversationType, cr.LastMessageId).Scan(&c.ConversationId, &c.CreatorID, &c.ConversationType, &c.LastMessageId)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
	}
	// Inert the conversation recipients
	_, err = tx.Exec(ctx, `INSERT INTO users_conversations (conversation_id, user_id) VALUES($1, $2)`, c.ConversationId, cr.RecipientId)
	if err != nil {
		return nil, apierror.DatabaseErrorClassification(path, op, err)
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
		c.last_message_id,
		c.created_at
        FROM conversations c
        WHERE c.conversation_id = $1
        AND (
            c.creator_id = $2
            OR EXISTS (
                SELECT 1 FROM users_conversations uc 
                WHERE uc.conversation_id = c.conversation_id AND uc.user_id = $2
            )
        )`, conversationId, userId).Scan(
		&c.ConversationId,
		&c.CreatorID,
		&c.ConversationType,
		&c.LastMessageId,
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
	    c.last_message_id, 
	    c.created_at,
	    (
		SELECT json_agg(
		    json_build_object(
			'id',		u.user_id::TEXT,
			'displayName',  u.username,
			'email',	u.email,
			'joinedDate',	(m.joined_at AT TIME ZONE 'UTC'),
			'role',		m.role
		    )
		)
		FROM (
		    SELECT 
			user_id, 
			joined_at, 
			'member' AS role 
		    FROM users_conversations 
		    WHERE conversation_id = c.conversation_id

		    UNION ALL

		    SELECT 
			creator_id AS user_id, 
			created_at AS joined_at, 
			'owner' AS role 
	FROM conversations 
		    WHERE conversation_id = c.conversation_id
		) AS m
		JOIN users u ON u.user_id = m.user_id
	    ) AS participants
	FROM conversations c
	WHERE c.creator_id = $1 
	   OR EXISTS (
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
			&c.LastMessageId,
			&c.CreatedAt,
			&participantsRaw,
		)

		if err != nil {
			return conversations, apierror.DatabaseErrorClassification(path, op, err)
		}

		if err := json.Unmarshal(participantsRaw, &c.Participants); err != nil {
			fmt.Println(err)
			return conversations, errors.B(path, op, "failed to unmarshal participants")
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
	err := r.db.QueryRow(ctx, `SELECT EXISTS (
            SELECT 1 
            FROM conversations c 
            WHERE c.conversation_id = $1 
            AND (
                c.creator_id = $2 
                OR EXISTS (
                    SELECT 1 FROM users_conversations uc 
                    WHERE uc.conversation_id = c.conversation_id AND uc.user_id = $2
                )
            )
        )`, conversationId, userId).Scan(&isExist)
	if err != nil {
		return pl, apierror.DatabaseErrorClassification(path, op, err)
	}

	// EXISTS returns a boolean value
	// true if there was a row or false if there was not
	// that's why we have to explicitly specifiy the error if there was no record.
	if !isExist {
		return pl, errors.B(path, op, errors.NotFound)
	}
	// Fetch the conversation participants.
	p, err := r.FetchMembers(ctx, conversationId)
	if err != nil {
		return pl, errors.B(path, op, err)
	}
	pl.Value = p
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
		list.joined_at,
		list.role
	    FROM users u
	    INNER JOIN (
		SELECT 
		    user_id, 
		    joined_at, 
		    'member' AS role 
		FROM users_conversations 
		WHERE conversation_id = $1

		UNION

		SELECT
		    c.creator_id, 
		    c.created_at AS joined_at, 
		    'owner' AS role 
		FROM conversations c 
		WHERE c.conversation_id = $1
	    ) AS list ON u.user_id = list.user_id;
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
