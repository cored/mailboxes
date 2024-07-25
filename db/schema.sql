-- Create mailboxes table
CREATE TABLE mailboxes (
		id INTEGER PRIMARY KEY,
		mpi_id VARCHAR(200),
		token VARCHAR(200),
		created_at TIMESTAMP
);

-- Create users table
CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		mailbox_id INTEGER,
		user_name VARCHAR(200),
		email_address VARCHAR(200),
		created_at TIMESTAMP,
		FOREIGN KEY (mailbox_id) REFERENCES mailboxes(id)
);

-- Insert sample data into mailboxes table
INSERT INTO mailboxes (id, mpi_id, token, created_at)
VALUES
		(1, 'mpi123', 'token123', '2024-07-23 12:00:00'),
		(2, 'mpi456', 'token456', '2024-07-23 13:00:00');

-- Insert sample data into users table
INSERT INTO users (id, mailbox_id, user_name, email_address, created_at)
VALUES
		(101, 1, 'user1', 'user1@example.com', '2024-07-23 12:30:00'),
		(102, 1, 'user2', 'user2@example.com', '2024-07-23 12:45:00'),
		(201, 2, 'user3', 'user3@example.com', '2024-07-23 13:15:00');
