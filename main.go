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

	mailboxChan, err := store.AllMailboxes()
	if err != nil {
		log.Fatalf("Error retrieving mailboxes: %v", err)
	}

	for mb := range mailboxChan {
		wg.Add(1)
		log.Printf("Processing %d mailbox", mb.ID)

		userChan, err := store.UsersForMailbox(mb.ID)
		if err != nil {
			log.Printf("Error retrieving users for mailbox %d: %v", mb.ID, err)
			wg.Done()
			continue
		}

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
	configPath := filepath.Join(".", "config.yaml")
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	dbDriver := viper.GetString("database.driver")
	dbPath := viper.GetString("database.path")

	store, err := db.NewDBStore(dbDriver, dbPath)
	if err != nil {
		log.Fatalf("Error setting up store: %v", err)
	}

	Pipeline(store)
}
