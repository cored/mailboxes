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

// Pipeline function to process mailboxes, retrieve users, and process each user
func Pipeline(store db.Store) {
	var wg sync.WaitGroup

	// Retrieve mailboxes directly from the store
	mailboxChan, err := store.AllMailboxes()
	if err != nil {
		log.Fatalf("Error retrieving mailboxes: %v", err)
	}

	for mb := range mailboxChan {
		wg.Add(1)
		log.Printf("Processing %d mailbox", mb.ID)

		// Retrieve users for each mailbox directly from the store
		userChan, err := store.UsersForMailbox(mb.ID)
		if err != nil {
			log.Printf("Error retrieving users for mailbox %d: %v", mb.ID, err)
			wg.Done()
			continue
		}

		// Launch a goroutine to process users for each mailbox
		go func(mb db.Mailbox) {
			defer wg.Done()

			userCount := 0
			for user := range userChan {
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
