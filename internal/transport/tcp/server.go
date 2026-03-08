package tcp

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/protocol"
	"github.com/iLeoon/realtime-gateway/internal/protocol/packets"
	"github.com/iLeoon/realtime-gateway/internal/transport/tcp/worker"
	"github.com/iLeoon/realtime-gateway/pkg/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

// server represents the central processing engine of the system. It is
// responsible for managing active WebSocket clients, receiving packets from
// the gateway, applying server-side logic, and routing messages to the
// appropriate clients.
type server struct {
	conf              *config.Config
	db                *pgxpool.Pool
	connections       map[uint32]net.Conn            // connectionID  → raw TCP conn
	clients           map[uint32][]uint32            // userID        → []connectionID
	userConversations map[uint32]map[uint32]struct{} // userID        → set of conversationIDs
	roomManager       map[uint32]map[uint32]struct{} // conversationID → set of memberIDs
	ready             chan<- struct{}
	mu                sync.RWMutex
	messagesCh        chan worker.Message
	done              chan struct{}
}

// MemberShip represents the rows returned from a DB query
type MemberShip struct {
	conversationID uint32
	memberID       uint32
}

// FanOut sends the messages to all the users within a conversation
type FanOut struct {
	rawConn      net.Conn
	connectionID uint32
	userID       uint32
}

// Create a new instance of the TCP server.
func Start(c *config.Config, db *pgxpool.Pool, ready chan<- struct{}) {
	server := &server{
		conf:              c,
		db:                db,
		connections:       make(map[uint32]net.Conn),
		clients:           make(map[uint32][]uint32),
		userConversations: make(map[uint32]map[uint32]struct{}),
		roomManager:       make(map[uint32]map[uint32]struct{}),
		ready:             ready,
		done:              make(chan struct{}, 1),
		messagesCh:        make(chan worker.Message, 200),
	}

	worker.New(server.done, server.messagesCh, server.db)
	server.listen()
}

// Lanunches the server, this method must be invoked inside a separate
// goroutine because it blocks while listening for incoming packets.
func (s *server) listen() {
	listner, err := net.Listen("tcp", s.conf.TcpPort)
	if err != nil {
		log.Error.Fatal("an error occured on creating tcp server", err)
		os.Exit(1)
	}
	log.Info.Println("TCP server is up and running...")
	defer listner.Close()

	close(s.ready)

	// Listening to the connections
	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Error.Println("error on trying to connect a tcp client", err)
			continue
		}
		go s.handleConn(conn)
	}
}

// handleConn is the main packet dispatcher for the TcpServer. It receives
// decoded packets from the gateway and routes them to their corresponding
// handlers using a switch on the packet opcode.
//
// It uses type assertion to convert the generic BuildPayload
// its concrete SendMessagePacket type.
func (s *server) handleConn(conn net.Conn) {
	var userID uint32
	const op errors.Op = "server.handleConn"
	const path errors.PathName = "tcp/server"
	stopPing := make(chan struct{})
	defer func() {
		close(stopPing)
		log.Info.Printf("%q: %q: tcp server terminated it's connection", path, op)
		conn.Close()
	}()
	go s.pingReq(conn, stopPing)

	for {
		if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			wrappedErr := errors.B(path, op, err)
			log.Error.Println("failed to read the packet from the peer", wrappedErr)
			return
		}
		// Call the decoder function on the connection to read
		// the incoming raw bytes and return the actual human-readable frame.
		frame, err := protocol.DecodeFrame(conn)
		if err != nil {
			// Error that gives context to the rest of the logs
			wrappedErr := errors.B(path, op, err)
			switch {
			case errors.Is(err, io.EOF):
				log.Info.Println("tcp connection is closed by peer(gateway)", wrappedErr)
				return
			case errors.Is(err, net.ErrClosed):
				log.Info.Println("connection is already terminated", err)
				return
			case errors.Is(err, os.ErrDeadlineExceeded):
				log.Info.Println("read deadline exceeded, closing connection", wrappedErr)
				return
			case errors.Is(err, errors.Client):
				if err := s.writePacket(&packets.ErrorPacket{
					Code:    errors.Client,
					Message: "invalid packet",
				}, conn); err != nil {
					log.Error.Println("failed to write the error packet to the gateway", err)
					return
				}
				return
			default:
				log.Error.Println("unexpected error while tcp server reading packets", wrappedErr)
				return
			}
		}

		// Route the frame to it's appropriate handler.
		// uses a type assertion to convert the generic BuildPayload interface into
		// its concrete *packet type
		switch p := frame.Payload.(type) {
		case *packets.ConnectPacket:
			userID = p.UserID
			err := s.register(p, conn)
			if err != nil {
				log.Error.Println("proccessing connect packet failed", err)
				s.handlePacketError(err, conn)
				return
			}
		case *packets.DisconnectPacket:
			log.Info.Println("Decode packet", "packet", p.String())
			s.unregister(p, conn)
			return
		case *packets.SendMessagePacket:
			err := s.handleSendMessageReq(p, userID)
			if err != nil {
				log.Error.Println("processing send message packet", err)
				s.handlePacketError(err, conn)
				return
			}
		case *packets.PongPacket:
			// We don't do anything we just renter the loop.
		case *packets.UpdateMessagePacket:
			err := s.handleUpdateMessagePacket(p, userID)
			if err != nil {
				log.Error.Println("processing update message packet failed", err)
				s.handlePacketError(err, conn)
				return
			}
		case *packets.DeleteMessagePacket:
			err := s.handleDeleteMessagePacket(p, userID)
			if err != nil {
				log.Error.Println("processing update message packet failed", err)
				s.handlePacketError(err, conn)
				return
			}

		case *packets.TypingPacket:
			err := s.handleTypingPacket(p, userID)
			if err != nil {
				log.Error.Println("processing typing packet failed", err)
				s.handlePacketError(err, conn)
				return
			}
		default:
			log.Error.Printf("invalid packet type from gateway: %T", p)
			return

		}
		log.Info.Println("Decode packet", "packet", frame.Payload.String())
	}
}

func (s *server) writePacket(pkt packets.BuildPayload, conn net.Conn) error {
	const path errors.PathName = "tcp/server"
	const op errors.Op = "server.writePacket"
	if err := conn.SetWriteDeadline(time.Now().Add(writeDuration)); err != nil {
		return errors.B(path, op, "connection is unhealty", err, errors.Network)
	}

	// Construct the frame, encode it, and then send it to the TCP client.
	frame := protocol.ConstructFrame(pkt)
	err := frame.EncodeFrame(conn)
	if err != nil {
		return errors.B(path, op, errors.Internal, err)
	}
	return nil
}

// handleSendMessage processes an inbound SendMessage packet
// it fans-out the messages to all the participants in a single conversation
// wether it was direct or group conversation.
func (s *server) handleSendMessageReq(pkt *packets.SendMessagePacket, userID uint32) error {
	const path errors.PathName = "tcp/server"
	const op errors.Op = "server.handleSendMessageReq"

	if userID == 0 {
		return errors.B(path, op, errors.Client, "userID is nonexistent")
	}

	if allowed := s.isAllowed(userID, pkt.ConversationID); !allowed {
		return errors.B(path, op, errors.Client, fmt.Errorf("the userID %v is not allowed to send messages in conversationID %v", userID, pkt.ConversationID))
	}

	// Pre-fetch the next message_id from Postgres sequence before fan-out.
	// This is a lightweight counter read (no table lock, no write) that lets us
	// include the real DB-assigned ID in the ResponseMessagePacket immediately,
	// while the actual INSERT remains async in the worker pool.
	var messageID uint32
	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := s.db.QueryRow(ctx, `SELECT nextval(pg_get_serial_sequence('messages', 'message_id'))`).Scan(&messageID)
		cancel()
		if err != nil {
			return errors.B(path, op, errors.Internal, fmt.Errorf("failed to pre-fetch message_id sequence: %w", err))
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// room is the a map that holds all the online users in the packet conversationID
	room, ok := s.roomManager[pkt.ConversationID]
	if !ok {
		return errors.B(path, op, errors.Client, fmt.Errorf("conversationID %v doesn't exit", pkt.ConversationID))
	}

	var fanOutTo []FanOut

	for m := range room {
		// members is all the users inside a conversation
		members := s.clients[m]
		for _, v := range members {
			fanOutTo = append(fanOutTo, FanOut{s.connections[v], v, m})
		}
	}

	for _, item := range fanOutTo {
		resPkt := &packets.ResponseMessagePacket{
			AuthorID:       userID,
			ConversationID: pkt.ConversationID,
			MessageID:      messageID,
			ResContent:     pkt.Content,
		}
		if err := s.writePacket(resPkt, item.rawConn); err != nil {
			writeErr := errors.B(path, op, err)
			log.Error.Printf("failed to send to connection: %d due to: %v", item.connectionID, writeErr)
		}
	}

	s.batchMessages(worker.Message{
		ID:             messageID,
		AuthorID:       userID,
		ConversationID: pkt.ConversationID,
		Content:        pkt.Content,
		Task:           worker.Insert,
	})

	return nil
}

// handleUpdateMessagePacket is responsible for updating a message
func (s *server) handleUpdateMessagePacket(pkt *packets.UpdateMessagePacket, userID uint32) error {
	const path errors.PathName = "tcp/server"
	const op errors.Op = "server.handleUpdateMessagePacket"

	if userID == 0 {
		return errors.B(path, op, errors.Client, "userID is nonexistent")
	}
	if allowed := s.isAllowed(userID, pkt.ConversationID); !allowed {
		return errors.B(path, op, errors.Client, fmt.Errorf("the userID %v is not allowed to send messages in conversationID %v", userID, pkt.ConversationID))
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	room, ok := s.roomManager[pkt.ConversationID]
	if !ok {
		return errors.B(path, op, errors.Client, fmt.Errorf("conversationID %v doesn't exit", pkt.ConversationID))
	}

	var fanOutTo []FanOut

	for m := range room {
		members := s.clients[m]
		for _, v := range members {
			fanOutTo = append(fanOutTo, FanOut{s.connections[v], v, m})
		}
	}

	now := time.Now().UTC()
	for _, item := range fanOutTo {
		resPkt := &packets.ResponseUpdateMessagePacket{
			MessageID:      pkt.MessageID,
			ConversationID: pkt.ConversationID,
			Updated_at:     now,
			ResContent:     pkt.Content,
		}
		if err := s.writePacket(resPkt, item.rawConn); err != nil {
			writeErr := errors.B(path, op, err)
			log.Error.Printf("failed to update the message of connection: %d due to: %v", item.connectionID, writeErr)
		}
	}

	s.batchMessages(worker.Message{
		ID:             pkt.MessageID,
		ConversationID: pkt.ConversationID,
		Content:        pkt.Content,
		UpdatedAt:      now,
		Task:           worker.Update,
	})

	return nil
}

// handleUpdateMessagePacket is responsible for deleteing a message
func (s *server) handleDeleteMessagePacket(pkt *packets.DeleteMessagePacket, userID uint32) error {
	const path errors.PathName = "tcp/server"
	const op errors.Op = "server.handleDeleteMessagePacket"

	if userID == 0 {
		return errors.B(path, op, errors.Client, "userID is nonexistent")
	}
	if allowed := s.isAllowed(userID, pkt.ConversationID); !allowed {
		return errors.B(path, op, errors.Client, fmt.Errorf("the userID %v is not allowed to send messages in conversationID %v", userID, pkt.ConversationID))
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	room, ok := s.roomManager[pkt.ConversationID]
	if !ok {
		return errors.B(path, op, errors.Client, fmt.Errorf("conversationID %v doesn't exit", pkt.ConversationID))
	}

	var fanOutTo []FanOut

	for m := range room {
		members := s.clients[m]
		for _, v := range members {
			fanOutTo = append(fanOutTo, FanOut{s.connections[v], v, m})
		}
	}

	for _, item := range fanOutTo {
		resPkt := &packets.ResponseDeleteMessagePacket{
			MessageID:      pkt.MessageID,
			ConversationID: pkt.ConversationID,
			AuthorID:       userID,
		}
		if err := s.writePacket(resPkt, item.rawConn); err != nil {
			writeErr := errors.B(path, op, err)
			log.Error.Printf("failed to delete the message of connection: %d due to: %v", item.connectionID, writeErr)
		}
	}

	s.batchMessages(worker.Message{
		ID:             pkt.MessageID,
		ConversationID: pkt.ConversationID,
		Task:           worker.Delete,
	})

	return nil
}

// handleTypingPacket fans out a typing indicator to all online members of
// the conversation.
func (s *server) handleTypingPacket(pkt *packets.TypingPacket, userID uint32) error {
	const path errors.PathName = "tcp/server"
	const op errors.Op = "server.handleTypingPacket"

	if userID == 0 {
		return errors.B(path, op, errors.Client, "userID is nonexistent")
	}

	if allowed := s.isAllowed(userID, pkt.ConversationID); !allowed {
		return errors.B(path, op, errors.Client, fmt.Errorf("userID %v is not a member of conversationID %v", userID, pkt.ConversationID))
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	room, ok := s.roomManager[pkt.ConversationID]
	if !ok {
		return errors.B(path, op, errors.Client, fmt.Errorf("conversationID %v doesn't exist", pkt.ConversationID))
	}

	resPkt := &packets.ResponseTypingPacket{
		ConversationID: pkt.ConversationID,
		UserID:         userID,
		IsTyping:       pkt.IsTyping,
	}

	for memberID := range room {
		if memberID == userID {
			continue
		}
		for _, connID := range s.clients[memberID] {
			conn := s.connections[connID]
			if err := s.writePacket(resPkt, conn); err != nil {
				log.Error.Printf("failed to send typing indicator to connectionID %d: %v", connID, err)
			}
		}
	}

	return nil
}

func (s *server) isAllowed(userID uint32, conversationID uint32) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if convs, ok := s.userConversations[userID]; ok {
		_, allowed := convs[conversationID]
		return allowed
	}
	return false
}

func (s *server) pingReq(conn net.Conn, stop <-chan struct{}) {
	const path errors.PathName = "tcp/server"
	const op errors.Op = "server.pingReq"
	ticker := time.NewTicker(pingTime)
	pkt := &packets.PingPacket{}
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.writePacket(pkt, conn); err != nil {
				errWrapper := errors.B(path, op, err)
				log.Error.Println("failed to send the ping packet, closing the connection:", errWrapper)
				conn.Close()
				return
			}
		case <-stop:
			return
		case <-s.done:
			return
		}
	}
}

// registerConnectionIDs add a connecteionIDs and userIDs to their maps.
func (s *server) register(pkt *packets.ConnectPacket, conn net.Conn) error {
	const path errors.PathName = "tcp/server"
	const op errors.Op = "server.register"

	memberships, err := s.fetchMemberships(pkt.UserID)
	if err != nil {
		return errors.B(path, op, err)
	}

	s.mu.Lock()
	// To prevent overwriting existing connections.
	_, ok := s.connections[pkt.ConnectionID]
	if ok {
		s.mu.Unlock()
		if err := s.writePacket(&packets.ErrorPacket{Code: errors.Client, Message: "connection already exists in the entry"}, conn); err != nil {
			return errors.B(path, op, err)
		}
		return errors.B(path, op, errors.Client, errors.Errorf("ConnectionID: %d already exists in the map", pkt.ConnectionID))
	}

	// Set the raw connection for every connectionID
	s.connections[pkt.ConnectionID] = conn
	// Map every connectionID to their authuenticated userID
	s.clients[pkt.UserID] = append(s.clients[pkt.UserID], pkt.ConnectionID)

	// Populate Map2 and Map3.
	if _, ok := s.userConversations[pkt.UserID]; !ok {
		s.userConversations[pkt.UserID] = make(map[uint32]struct{})
	}
	for _, m := range memberships {
		s.userConversations[pkt.UserID][m.conversationID] = struct{}{}
		if _, ok := s.roomManager[m.conversationID]; !ok {
			s.roomManager[m.conversationID] = make(map[uint32]struct{})
		}
		s.roomManager[m.conversationID][m.memberID] = struct{}{}
	}

	// Collect presence targets while still under write lock.
	// Only fan-out online on the first connection for this user.
	var onlineTargets []net.Conn
	if len(s.clients[pkt.UserID]) == 1 {
		for convID := range s.userConversations[pkt.UserID] {
			for memberID := range s.roomManager[convID] {
				if memberID == pkt.UserID {
					continue
				}
				for _, connID := range s.clients[memberID] {
					onlineTargets = append(onlineTargets, s.connections[connID])
				}
			}
		}
	}
	s.mu.Unlock()

	if len(onlineTargets) > 0 {
		resPkt := &packets.ResponsePresencePacket{UserID: pkt.UserID, IsOnline: true}
		log.Info.Println("Decode packet", "packet", resPkt.String())

		for _, conn := range onlineTargets {
			if err := s.writePacket(resPkt, conn); err != nil {
				log.Error.Printf("failed to send online presence for userID %d: %v", pkt.UserID, err)
			}
		}
	}

	return nil
}

// unRegisterConnectionIDs removes the connectionIDs and userIDs from the map
func (s *server) unregister(pkt *packets.DisconnectPacket, conn net.Conn) {
	s.mu.Lock()

	delete(s.connections, pkt.ConnectionID)

	clients := s.clients[pkt.UserID]
	filtered := make([]uint32, 0, len(clients))
	for _, client := range clients {
		if client != pkt.ConnectionID {
			filtered = append(filtered, client)
		}
	}

	var offlineTargets []net.Conn
	if len(filtered) == 0 {
		// Last connection for this user — collect presence targets before cleanup.
		for convID := range s.userConversations[pkt.UserID] {
			for memberID := range s.roomManager[convID] {
				if memberID == pkt.UserID {
					continue
				}
				for _, connID := range s.clients[memberID] {
					offlineTargets = append(offlineTargets, s.connections[connID])
				}
			}
		}

		delete(s.clients, pkt.UserID)
		for convID := range s.userConversations[pkt.UserID] {
			delete(s.roomManager[convID], pkt.UserID)
			if len(s.roomManager[convID]) == 0 {
				delete(s.roomManager, convID)
			}
		}
		delete(s.userConversations, pkt.UserID)
	} else {
		s.clients[pkt.UserID] = filtered
	}

	s.mu.Unlock()

	if len(offlineTargets) > 0 {
		resPkt := &packets.ResponsePresencePacket{UserID: pkt.UserID, IsOnline: false}
		log.Info.Println("Decode packet", "packet", resPkt.String())

		for _, conn := range offlineTargets {
			if err := s.writePacket(resPkt, conn); err != nil {
				log.Error.Printf("failed to send offline presence for userID %d: %v", pkt.UserID, err)
			}
		}
	}
}

func (s *server) batchMessages(message worker.Message) {
	const op errors.Op = "server.batchMessage"
	const path errors.PathName = "tcp/server"
	select {
	case s.messagesCh <- message:
	case <-s.done:
	default:
		// The channel is full and can't accept more values
		log.Error.Println(errors.B(path, op, errors.ServiceUnavailable, "message dropped: worker buffer full"))
	}
}

func (s *server) handlePacketError(err error, conn net.Conn) {
	if errors.Is(err, errors.Client) {
		errWrite := s.writePacket(&packets.ErrorPacket{
			Code:    errors.Client,
			Message: "unexpected error try refreshing the page",
		}, conn)

		if errWrite != nil {
			log.Error.Println("failed to send 'error packet' alert to client:", errWrite)
		}
	}
}

// fetchMemberships queries every conversation the user belongs to and all
// members of those conversations in a single round-trip. Must be called
// outside the mutex — it performs I/O.
func (s *server) fetchMemberships(userID uint32) ([]MemberShip, error) {
	const path errors.PathName = "tcp/server"
	const op errors.Op = "server.fetchMemberships"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.Query(ctx, `
		SELECT uc1.conversation_id, uc2.user_id
		FROM users_conversations uc1
		JOIN users_conversations uc2 ON uc1.conversation_id = uc2.conversation_id
		WHERE uc1.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, errors.B(path, op, errors.Internal, err)
	}
	defer rows.Close()

	var memberships []MemberShip
	for rows.Next() {
		var m MemberShip
		if err := rows.Scan(&m.conversationID, &m.memberID); err != nil {
			return nil, errors.B(path, op, errors.Internal, err)
		}
		memberships = append(memberships, m)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.B(path, op, errors.Internal, err)
	}
	return memberships, nil
}
