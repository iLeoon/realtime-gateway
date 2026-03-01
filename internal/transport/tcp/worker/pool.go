package worker

import (
	// "runtime"

	"context"
	"time"

	"github.com/iLeoon/realtime-gateway/pkg/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	AuthorID       uint32
	ConversationID uint32
	Content        string
}

// New startes a worker pool
func New(done <-chan struct{}, messagesCh <-chan Message, db *pgxpool.Pool) {
	// var workers = runtime.NumCPU() * 2
	for i := 0; i < 2; i++ {
		// Create worker
		go workers(db, done, messagesCh)
	}
}

func workers(db *pgxpool.Pool, done <-chan struct{}, messagesCh <-chan Message) {
	for {
		select {
		case m, ok := <-messagesCh:
			if !ok {
				return
			}
			store(db, m)
		case <-done:
			return
		}
	}
}

func store(db *pgxpool.Pool, message Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.Exec(ctx, `INSERT INTO messages (creator_id, conversation_id, content) VALUES ($1, $2, $3)`, message.AuthorID, message.ConversationID, message.Content)
	if err != nil {
		log.Error.Printf("failed to store message: %v", err)
	}
}
