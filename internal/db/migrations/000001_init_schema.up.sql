
CREATE TYPE conversation_type AS ENUM ('group-chat', 'private-chat');

COMMENT ON TYPE conversation_type IS
'Defines the kind of conversation. ';

CREATE TYPE friends_type AS ENUM ('accepted', 'rejected', 'pending');

COMMENT ON TYPE friends_type IS
'Represents the state of a friendship or friend request between two users.';

CREATE TABLE IF NOT EXISTS users (
	user_id INT GENERATED ALWAYS AS IDENTITY,
	username varchar(30) NOT NULL,
	email varchar(30) NOT NULL UNIQUE,
	PRIMARY KEY (user_id)
);

COMMENT ON TABLE users IS
'Stores application users. 
This table represents user identities only and does not store relationships or activity.';

CREATE TABLE IF NOT EXISTS conversations (
	conversation_id INT GENERATED ALWAYS AS IDENTITY,
	creator_id INT NOT NULL,
	conversation_type conversation_type NOT NULL,
	last_message_id INT NULL,
	created_at TIMESTAMP DEFAULT NOW(),

	PRIMARY KEY (conversation_id),
	FOREIGN KEY (creator_id) REFERENCES users (user_id)

);

COMMENT ON TABLE conversations IS
'Stores users conversations.
A conversation can be a private chat or a group chat.';

CREATE TABLE IF NOT EXISTS messages(
	message_id INT GENERATED ALWAYS AS IDENTITY,
	creator_id INT NOT NULL, 
	conversation_id INT NOT NULL,
	content TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT NOW(),
	deleted_at TIMESTAMP NULL,

	PRIMARY KEY (message_id),
	FOREIGN KEY (creator_id) REFERENCES users (user_id),
	FOREIGN KEY (conversation_id) REFERENCES conversations (conversation_id) ON DELETE CASCADE
);

COMMENT ON TABLE messages IS
'Stores messages sent inside conversations.
Messages belong to exactly one conversation and are sent by one user.';


CREATE TABLE IF NOT EXISTS users_conversations(
	conversation_id INT,
	user_id INT,
	joined_at TIMESTAMP DEFAULT NOW(),
	left_at TIMESTAMP NULL,
	
	PRIMARY KEY (conversation_id, user_id),
	FOREIGN KEY (conversation_id) REFERENCES conversations (conversation_id) ON DELETE CASCADE,
	FOREIGN KEY (user_id) REFERENCES users (user_id) 
);

COMMENT ON TABLE users_conversations IS
'Stores users in every conversations.
Used for both private chats and group chats.';



CREATE TABLE IF NOT EXISTS friends(
	sender_id INT,
	recipient_id INT,
	status friends_type NOT NULL,

	PRIMARY KEY (sender_id, recipient_id),
	FOREIGN KEY (sender_id) REFERENCES users (user_id),
	FOREIGN KEY (recipient_id) REFERENCES users (user_id),
	CHECK (sender_id <> recipient_id)
);

COMMENT ON TABLE friends IS
'Represents friendship relationships and friend requests between users.';


CREATE TABLE IF NOT EXISTS providers(
	provider TEXT NOT NULL DEFAULT 'google',
	provider_user_id TEXT NOT NULL,
	user_id INT NOT NULL,
	created_at TIMESTAMP DEFAULT NOW(),
	
	PRIMARY KEY (provider, provider_user_id),
	FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);

COMMENT ON TABLE providers IS
'Store needed providers data';

ALTER TABLE conversations 
	ADD CONSTRAINT fk_last_message FOREIGN KEY (last_message_id) REFERENCES messages (message_id);
