package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/pkg/log"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskType int8

const (
	Insert TaskType = iota
	Update
	Delete
)

type Message struct {
	ID             uint32
	AuthorID       uint32
	ConversationID uint32
	Content        string
	UpdatedAt      time.Time
	Task           TaskType
}

func New(done <-chan struct{}, messagesCh <-chan Message, db *pgxpool.Pool) {
	for i := 0; i < 2; i++ {
		go worker(db, done, messagesCh)
	}
}

func worker(db *pgxpool.Pool, done <-chan struct{}, messagesCh <-chan Message) {
	for {
		select {
		case m, ok := <-messagesCh:
			if !ok {
				return
			}
			switch m.Task {
			case Insert:
				if err := store(db, m); err != nil {
					log.Error.Printf("failed to persist message: %v", err)
				}

			case Update:
				if err := update(db, m); err != nil {
					log.Error.Printf("failed to update message: %v", err)
				}

			case Delete:
				if err := delete(db, m); err != nil {
					log.Error.Printf("failed to delete message: %v", err)
				}
			}

		case <-done:
			return
		}
	}
}

func store(db *pgxpool.Pool, m Message) error {
	const path errors.PathName = "worker/pool"
	const op errors.Op = "worker.store"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.Exec(ctx,
		`INSERT INTO messages (message_id, creator_id, conversation_id, content)
		 VALUES ($1, $2, $3, $4)`,
		m.ID, m.AuthorID, m.ConversationID, m.Content,
	)
	if err != nil {
		switch e := err.(type) {
		case *pgconn.PgError:
			switch e.Code {
			case "08006", "08001", "08003":
				return errors.B(path, op, errors.Network, err)
			case "53300":
				return errors.B(path, op, errors.ServiceUnavailable, err)
			case "57014":
				return errors.B(path, op, errors.TimeOut, err)
			default:
				return errors.B(path, op, errors.Internal, err)
			}
		}
		return errors.B(path, op, errors.Internal, err)
	}
	return nil
}

func update(db *pgxpool.Pool, m Message) error {
	const path errors.PathName = "worker/pool"
	const op errors.Op = "worker.update"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := db.Exec(ctx,
		`UPDATE messages m SET content = $1, edited_at = $2
		 WHERE m.message_id = $3 AND m.conversation_id = $4
		`,
		m.Content, m.UpdatedAt, m.ID, m.ConversationID)
	if err != nil {
		switch e := err.(type) {
		case *pgconn.PgError:
			switch e.Code {
			case "08006", "08001", "08003":
				return errors.B(path, op, errors.Network, err)
			case "53300":
				return errors.B(path, op, errors.ServiceUnavailable, err)
			case "57014":
				return errors.B(path, op, errors.TimeOut, err)
			default:
				return errors.B(path, op, errors.Internal, err)
			}
		}
		return errors.B(path, op, errors.Internal, err)
	}

	if tx.RowsAffected() == 0 {
		return errors.B(path, op, errors.Internal, fmt.Errorf("failed to update messageID: %v", m.ID))
	}

	return nil
}

func delete(db *pgxpool.Pool, m Message) error {
	const path errors.PathName = "worker/pool"
	const op errors.Op = "worker.delete"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := db.Exec(ctx,
		`UPDATE messages m SET deleted_at = now()
		 WHERE m.message_id = $1 AND m.conversation_id = $2
		`,
		m.ID, m.ConversationID)
	if err != nil {
		switch e := err.(type) {
		case *pgconn.PgError:
			switch e.Code {
			case "08006", "08001", "08003":
				return errors.B(path, op, errors.Network, err)
			case "53300":
				return errors.B(path, op, errors.ServiceUnavailable, err)
			case "57014":
				return errors.B(path, op, errors.TimeOut, err)
			default:
				return errors.B(path, op, errors.Internal, err)
			}
		}
		return errors.B(path, op, errors.Internal, err)
	}

	if tx.RowsAffected() == 0 {
		return errors.B(path, op, errors.Internal, fmt.Errorf("failed to delete messageID: %v", m.ID))
	}

	return nil
}
