package main

import (
	"database/sql"
	"path/filepath"
	"log"
	"sync"

	"github.com/spf13/viper"
	_ "github.com/mattn/go-sqlite3"
)

// Mailbox struct represents a mailbox from the database
type Mailbox struct {
	ID        int
	MPIID     string
	Token     string
	CreatedAt string // Assuming timestamp is read as string for simplicity
}

// User struct represents a user from the database
type User struct {
	ID           int
	MailboxID    int
	UserName     string
	EmailAddress string
	CreatedAt    string // Assuming timestamp is read as string for simplicity
}

// processUser is a fictional function to process each user
func processUser(user User) {
	log.Printf("Processing user: User Name - %s, Mailbox Token - %s", user.UserName, "<fake_token>")
}

// Function to retrieve mailboxes and return them via a channel
func RetrieveMailboxes(db *sql.DB) <-chan Mailbox {

	mailboxChannel := make(chan Mailbox, 100) // Buffered channel with capacity 100

	go func() {
		defer close(mailboxChannel)

		rows, err := db.Query("SELECT id, mpi_id, token, created_at FROM mailboxes")
		if err != nil {
			log.Printf("Error querying mailboxes: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var mailbox Mailbox
			err := rows.Scan(&mailbox.ID, &mailbox.MPIID, &mailbox.Token, &mailbox.CreatedAt)
			if err != nil {
				log.Printf("Error scanning mailbox row: %v", err)
				continue
			}
			mailboxChannel <- mailbox
		}

		if err := rows.Err(); err != nil {
			log.Printf("Error iterating over mailbox rows: %v", err)
			return
		}
	}()

	return mailboxChannel
}

// Function to retrieve users for a given mailbox ID and return them via a channel
func RetrieveUsersForMailbox(db *sql.DB, mailboxID int) <-chan User {
	userChannel := make(chan User, 100) // Buffered channel with capacity 100

	go func() {
		defer close(userChannel)

		query := "SELECT id, mailbox_id, user_name, email_address, created_at FROM users WHERE mailbox_id = ?"
		rows, err := db.Query(query, mailboxID)
		if err != nil {
			log.Printf("Error querying users for mailbox %d: %v", mailboxID, err)
			return
		}
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

	return userChannel
}

// Pipeline function to process mailboxes, retrieve users, and process each user
func Pipeline(db *sql.DB) {
	mailboxes := RetrieveMailboxes(db)
	var wg sync.WaitGroup

	for mailbox := range mailboxes {
		wg.Add(1)
		log.Printf("Processing %d mailbox", mailbox.ID)

		users := RetrieveUsersForMailbox(db, mailbox.ID)

		// Launch a goroutine to process users for each mailbox
		go func(mb Mailbox) {
			defer wg.Done()

			userCount := 0
			for user := range users {
				processUser(user)
				userCount++
			}

			log.Printf("%d users processed for mailbox %d", userCount, mb.ID)
		}(mailbox)
	}

	wg.Wait()
}


// Main function to initialize database connection and call the pipeline function
func main() {
	// Initialize viper for configuration
	configPath := filepath.Join(".", "config.yaml")
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Set up database connection based on configuration
	dbDriver := viper.GetString("database.driver")
	dbPath := viper.GetString("database.path")

	db, err := sql.Open(dbDriver, dbPath)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Call the pipeline function to process mailboxes and users
	Pipeline(db)
}
