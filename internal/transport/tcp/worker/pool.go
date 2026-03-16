package worker

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/pkg/log"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const path errors.PathName = "worker/pool"

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
	var workers = runtime.NumCPU()
	for i := 0; i < workers*2; i++ {
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
			handleTask(m.Task, db, m)
		case <-done:
			return
		}
	}
}

func handleTask(task TaskType, db *pgxpool.Pool, message Message) {
	var err error
	switch task {
	case Insert:
		err = store(db, message)
	case Update:
		err = update(db, message)
	case Delete:
		err = delete(db, message)
	}
	if err != nil {
		log.Error.Printf("failed to process %v message: %v", message, err)
	}
}

func store(db *pgxpool.Pool, m Message) error {
	const op errors.Op = "worker.store"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.Exec(ctx,
		`INSERT INTO messages (message_id, creator_id, conversation_id, content)
		 VALUES ($1, $2, $3, $4)`,
		m.ID, m.AuthorID, m.ConversationID, m.Content,
	)
	if err != nil {
		return handleDBError(err, op)
	}
	return nil
}

func update(db *pgxpool.Pool, m Message) error {
	const op errors.Op = "worker.update"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := db.Exec(ctx,
		`UPDATE messages m SET content = $1, edited_at = $2
		 WHERE m.message_id = $3 AND m.conversation_id = $4
		`,
		m.Content, m.UpdatedAt, m.ID, m.ConversationID)
	if err != nil {
		return handleDBError(err, op)
	}

	if tx.RowsAffected() == 0 {
		return errors.B(path, op, errors.Internal, fmt.Errorf("failed to update messageID: %v", m.ID))
	}

	return nil
}

func delete(db *pgxpool.Pool, m Message) error {
	const op errors.Op = "worker.delete"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := db.Exec(ctx,
		`UPDATE messages m SET deleted_at = now()
		 WHERE m.message_id = $1 AND m.conversation_id = $2
		`,
		m.ID, m.ConversationID)
	if err != nil {
		return handleDBError(err, op)
	}

	if tx.RowsAffected() == 0 {
		return errors.B(path, op, errors.Internal, fmt.Errorf("failed to delete messageID: %v", m.ID))
	}

	return nil
}

func handleDBError(err error, op errors.Op) error {
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
