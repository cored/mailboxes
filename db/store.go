
package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

// SQLiteStore implements the Store interface using SQLite
type DBStore struct {
	db *sql.DB
	log *log.Logger
}

// NewDBStore initializes a new DBStore instance.
func NewDBStore(dbDriver, dbSource string) (Store, error) {
	db, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return nil, err
	}
	return &DBStore{db: db, log: log.Default()}, nil
}

// AllMailboxes retrieves all mailboxes from the database using channels and goroutines.
func (s *DBStore) AllMailboxes() (<-chan Mailbox, error) {
	query := "SELECT id, mpi_id, token, created_at FROM mailboxes"

	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("Error querying mailboxes: %v", err)
		return nil, err
	}

	mailboxChannel := make(chan Mailbox)

	go func() {
		defer close(mailboxChannel)
		defer rows.Close()

		for rows.Next() {
			var mb Mailbox
			err := rows.Scan(&mb.ID, &mb.MPIID, &mb.Token, &mb.CreatedAt)
			if err != nil {
				log.Printf("Error scanning mailbox row: %v", err)
				continue
			}
			mailboxChannel <- mb
		}

		if err := rows.Err(); err != nil {
			log.Printf("Error iterating over mailbox rows: %v", err)
			return
		}
	}()

	return mailboxChannel, nil
}

// UsersForMailbox retrieves all users for a given mailbox ID from the database using channels and goroutines.
func (s *DBStore) UsersForMailbox(mailboxID int) (<-chan User, error) {
	query := "SELECT id, mailbox_id, user_name, email_address, created_at FROM users WHERE mailbox_id = ?"

	rows, err := s.db.Query(query, mailboxID)
	if err != nil {
		log.Printf("Error querying users for mailbox %d: %v", mailboxID, err)
		return nil, err
	}

	userChannel := make(chan User)

	go func() {
		defer close(userChannel)
		defer rows.Close()

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

	return userChannel, nil
}
