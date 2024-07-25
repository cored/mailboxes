package db

import (
	"database/sql"
	"log"
)

// SQLiteStore implements the Store interface using SQLite
type DBStore struct {
	db *sql.DB
}

func NewDBStore(dbDriver, dbSource string) (Store, error) {
	db, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return nil, err
	}
	return &DBStore{db: db}, nil
}

// AllMailboxes retrieves all mailboxes from the database
func (s *DBStore) AllMailboxes() ([]Mailbox, error) {
	query := "SELECT id, mpi_id, token, created_at FROM mailboxes"

	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("Error querying mailboxes: %v", err)
		return nil, err
	}
	defer rows.Close()

	var mailboxes []Mailbox
	for rows.Next() {
		var mb Mailbox
		err := rows.Scan(&mb.ID, &mb.MPIID, &mb.Token, &mb.CreatedAt)
		if err != nil {
			log.Printf("Error scanning mailbox row: %v", err)
			continue
		}
		mailboxes = append(mailboxes, mb)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over mailbox rows: %v", err)
		return nil, err
	}

	return mailboxes, nil
}


// UsersForMailbox retrieves all users for a given mailbox ID from the database using channels
func (s *DBStore) UsersForMailbox(mailboxID int) ([]User, error) {
	query := "SELECT id, mailbox_id, user_name, email_address, created_at FROM users WHERE mailbox_id = ?"

	rows, err := s.db.Query(query, mailboxID)
	if err != nil {
		log.Printf("Error querying users for mailbox %d: %v", mailboxID, err)
		return nil, err
	}
	defer rows.Close()

	// Channel to receive users asynchronously
	userChannel := make(chan User, 100) // Buffered channel with capacity 100

	// Concurrently fetch users and send them to the channel
	go func() {
		defer close(userChannel)

		for rows.Next() {
			var user User
			err := rows.Scan(&user.ID, &user.MailboxID, &user.UserName, &user.EmailAddress, &user.CreatedAt)
			if err != nil {
				log.Printf("Error scanning user row: %v", err)
				continue
			}
			userChannel <- user
		}

		if err := rows.Err(); err != nil {
			log.Printf("Error iterating over user rows: %v", err)
			return
		}
	}()

	users := collectUsers(userChannel)

	return users, nil
}

// Collect users from the channel into a slice
func collectUsers(userChannel <-chan User) []User {
	var users []User
	for user := range userChannel {
		users = append(users, user)
	}
	return users
}
