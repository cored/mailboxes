# Writing a Data Pipeline in Go for Beginners

## Introduction

A data pipeline is a series of steps that process data from a source to a destination. We’ll build a basic pipeline to fetch mailboxes from a database, retrieve users for each mailbox, and process those users. We’ll use Go’s concurrency features, channels, and goroutines to handle data processing efficiently.

## Prerequisites

Before you start, ensure you have:

- Basic knowledge of Go programming
- Go installed on your system
- A text editor or IDE for Go development

## Project Setup

### 1. Create the Project Structure

Start by creating a new directory for your project:

```bash
mkdir data-pipeline
cd data-pipeline
```

Create the following files and directories:

```bash
touch main.go
touch db.go
touch main_test.go
mkdir config
touch config/config.yaml
```

### 2. Define the Database Schema

For our example, we’ll use an SQLite database. Define the schema with two tables: `mailboxes` and `users`.

**Create a SQL script (`setup.sql`) for initializing the database:**

```sql
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
```

### 3. Implement the Database Access Layer

In `db.go`, define the structures and methods to interact with the database.

**`db.go`**

```go
package db

import (
	"database/sql"
	"log"
)

// Mailbox represents a mailbox record
type Mailbox struct {
	ID        int
	MPIID     string
	Token     string
	CreatedAt string
}

// User represents a user record
type User struct {
	ID           int
	MailboxID    int
	UserName     string
	EmailAddress string
	CreatedAt    string
}

// Store interface for accessing data
type Store interface {
	AllMailboxes() (<-chan Mailbox, error)
	UsersForMailbox(mailboxID int) (<-chan User, error)
}

// DBStore implements the Store interface using SQLite
type DBStore struct {
	db *sql.DB
}

// NewDBStore initializes a new DBStore instance
func NewDBStore(dbDriver, dbSource string) (Store, error) {
	db, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return nil, err
	}
	return &DBStore{db: db}, nil
}

// AllMailboxes retrieves all mailboxes from the database
func (s *DBStore) AllMailboxes() (<-chan Mailbox, error) {
	query := "SELECT id, mpi_id, token, created_at FROM mailboxes"
	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("Error querying mailboxes: %v", err)
		return nil, err
	}

	mailboxChan := make(chan Mailbox)
	go func() {
		defer close(mailboxChan)
		defer rows.Close()

		for rows.Next() {
			var mb Mailbox
			err := rows.Scan(&mb.ID, &mb.MPIID, &mb.Token, &mb.CreatedAt)
			if err != nil {
				log.Printf("Error scanning mailbox row: %v", err)
				continue
			}
			mailboxChan <- mb
		}

		if err := rows.Err(); err != nil {
			log.Printf("Error iterating over mailbox rows: %v", err)
			return
		}
	}()

	return mailboxChan, nil
}

// UsersForMailbox retrieves all users for a given mailbox ID
func (s *DBStore) UsersForMailbox(mailboxID int) (<-chan User, error) {
	query := "SELECT id, mailbox_id, user_name, email_address, created_at FROM users WHERE mailbox_id = ?"
	rows, err := s.db.Query(query, mailboxID)
	if err != nil {
		log.Printf("Error querying users for mailbox %d: %v", mailboxID, err)
		return nil, err
	}

	userChan := make(chan User)
	go func() {
		defer close(userChan)
		defer rows.Close()

		for rows.Next() {
			var user User
			err := rows.Scan(&user.ID, &user.MailboxID, &user.UserName, &user.EmailAddress, &user.CreatedAt)
			if err != nil {
				log.Printf("Error scanning user row: %v", err)
				continue
			}
			userChan <- user
		}

		if err := rows.Err(); err != nil {
			log.Printf("Error iterating over user rows: %v", err)
			return
		}
	}()

	return userChan, nil
}
```

### 4. Implement the Data Pipeline

**`main.go`**

```go
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
```

### 5. Testing

**`main_test.go`**

```go
package main

import (
	"log"
	"testing"
	"time"
	"sync"

	"mailboxes/db"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// setupMockDB initializes a mock database connection and returns the database instance and sqlmock instance
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing mock database: %v", err)
	}
	return db, mock
}

// TestPipeline tests the entire pipeline using MockStore
func TestPipeline(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT id, mpi_id, token, created_at FROM mailboxes").
		WillReturnRows(sqlmock.NewRows([]string{"id", "mpi_id", "token", "created_at"}).
			AddRow(1, "mpi123", "token123", "2024-07-23 12:00:00").
			AddRow(2, "mpi456", "token456", "2024-07-23 13:00:00"))

	mock.ExpectQuery("SELECT id, mailbox_id, user_name, email_address, created_at FROM users WHERE mailbox_id = ?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "mailbox_id", "user_name", "email_address", "created_at"}).
			AddRow(101, 1, "user1", "user1@example.com", "2024-07-23 12:30:00").
			AddRow(102, 1, "user2", "user2@example.com", "2024-07-23 12:45:00"))

	mock.ExpectQuery("SELECT id, mailbox_id, user_name, email_address, created_at FROM users WHERE mailbox_id = ?").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "mailbox_id", "user_name", "email_address", "created_at"}).
			AddRow(201, 2, "user3", "user3@example.com", "2024-07-23 13:30:00").
			AddRow(202, 2, "user4", "user4@example.com", "2024-07-23 13:45:00"))

	store := &db.DB

Store{db: db}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		Pipeline(store)
	}()

	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
```

### 6. Running the Program

**Running the Program**

Ensure you have set up your `config.yaml` correctly. Then, run your Go program:

```bash
go run main.go
```

**Running the Tests**

To run your tests, use the following command:

```bash
go test -v
```

Congratulations! You’ve successfully created a simple data pipeline in Go. This pipeline fetches mailboxes from a database, retrieves users for each mailbox, and processes those users concurrently. You’ve also learned how to write tests for your pipeline using mock databases to ensure everything works correctly.

Happy coding!
