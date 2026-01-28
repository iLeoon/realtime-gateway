package conversation

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateOrFindConversation(ctx context.Context, creatorId int) (err error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateOrFindConversation(ctx context.Context, creatorId int) (err error) {
	// var conversationId int
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// err = tx.QueryRow(ctx, `INSERT INTO conversations (creator_id, conversation_type)
	// VALUES($1, private-chat) RETURNING conversation_id`,
	// creatorId).Scan(&conversationId)

	// tx.Exec(`INSERT INTO users_conversations`)

	return nil
}
