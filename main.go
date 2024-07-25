package main

import (
	"log"
	"path/filepath"
	"sync"

	"mailboxes/db" // Import the store package

	"github.com/spf13/viper"
)

// processUser is a fictional function to process each user
func processUser(user db.User) {
	log.Printf("Processing user: User Name - %s, Mailbox Token - %s", user.UserName, "<fake_token>")
}

// Function to retrieve mailboxes and return them via a channel
func RetrieveMailboxes(store db.Store) <-chan db.Mailbox {
	mailboxChannel := make(chan db.Mailbox) // Buffered channel with capacity 100

	go func() {
		defer close(mailboxChannel)

		mailboxes, err := store.AllMailboxes()
		if err != nil {
			log.Printf("Error retrieving mailboxes: %v", err)
			return
		}

		for _, mb := range mailboxes {
			mailboxChannel <- mb
		}
	}()

	return mailboxChannel
}

// Function to retrieve users for a given mailbox ID and return them via a channel
func RetrieveUsersForMailbox(store db.Store, mailboxID int) <-chan db.User {
	userChannel := make(chan db.User, 100) // Buffered channel with capacity 100

	go func() {
		defer close(userChannel)

		users, err := store.UsersForMailbox(mailboxID)
		if err != nil {
			log.Printf("Error retrieving users for mailbox %d: %v", mailboxID, err)
			return
		}

		for _, user := range users {
			userChannel <- user
		}
	}()

	return userChannel
}

// Pipeline function to process mailboxes, retrieve users, and process each user
func Pipeline(store db.Store) {
	mailboxes := RetrieveMailboxes(store)
	var wg sync.WaitGroup

	for mb := range mailboxes {
		wg.Add(1)
		log.Printf("Processing %d mailbox", mb.ID)

		users := RetrieveUsersForMailbox(store, mb.ID)

		// Launch a goroutine to process users for each mailbox
		go func(mb db.Mailbox) {
			defer wg.Done()

			userCount := 0
			for user := range users {
				processUser(user)
				userCount++
			}

			log.Printf("%d users processed for mailbox %d", userCount, mb.ID)
		}(mb)
	}

	wg.Wait()
}

// Main function to initialize configuration, database connection, and call the pipeline function
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

	store, err := db.NewDBStore(dbDriver, dbPath)
	if err != nil {
		log.Fatalf("Error setting up store: %v", err)
	}

	// Call the pipeline function to process mailboxes and users
	Pipeline(store)
}

