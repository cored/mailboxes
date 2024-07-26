## Writing a Data Pipeline in Go

### Introduction

In this tutorial, we will build a simple data pipeline in Go. Our goal is to retrieve data from a SQLite database, process it concurrently, and handle errors gracefully. We'll focus on separating data access logic from the pipeline logic, using specific dependencies for configuration, database interactions, and testing.

### Prerequisites

- **Go Programming Language**: Basic knowledge of Go syntax and concurrency.
- **SQLite**: Understanding of SQLite as a lightweight database.
- **Dependencies**: Familiarity with Go modules and package management.

### Dependencies

1. **`github.com/spf13/viper`**: For configuration management.
2. **`github.com/DATA-DOG/go-sqlmock`**: For mocking SQL queries in tests.
3. **`github.com/stretchr/testify`**: For assertions in tests (optional but recommended).

###	Overview

We will create a simple data pipeline that retrieves mailboxes and users from a SQLite database. The solution in support of the following user stories:

1. As a user, I want to retrieve all mailboxes from the database.
2. As a user, I wan to retrieve all users for a specific mailbox.

### Project Structure

Here is a suggested directory structure:

```
/project-root
	/db
		db.go
		db_test.go
	/main.go
	/config.yaml
	/scripts
		run.sh
```

### 1. Data Access Layer

#### `db/db.go`

This file contains our data access logic. We define interfaces and concrete implementations for interacting with the database.

**`db.go`**

```go
package db

import (
	"database/sql"
	"log"
)

// Store defines the methods required for accessing data.
type Store interface {
	AllMailboxes() (<-chan Mailbox, error)
	UsersForMailbox(mailboxID int) (<-chan User, error)
}

// DBStore implements the Store interface using SQLite.
type DBStore struct {
	db *sql.DB
}

// NewDBStore initializes a new DBStore instance.
func NewDBStore(dbDriver, dbSource string) (Store, error) {
	db, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return nil, err
	}
	return &DBStore{db: db}, nil
}

// AllMailboxes retrieves all mailboxes from the database.
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
		}
	}()

	return mailboxChannel, nil
}

// UsersForMailbox retrieves users for a specific mailbox ID.
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
		}
	}()

	return userChannel, nil
}

// Mailbox represents a mailbox entity.
type Mailbox struct {
	ID        int
	MPIID     string
	Token     string
	CreatedAt string
}

// User represents a user entity.
type User struct {
	ID           int
	MailboxID    int
	UserName     string
	EmailAddress string
	CreatedAt    string
}
```

### 2. Data Pipeline

#### `main.go`

This file contains the pipeline logic that uses the `DBStore` to process data.

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
```

### 3. Testing

Testing is crucial for verifying the correctness of your data pipeline. We use the `sqlmock` library to mock database interactions and verify that our data access methods behave as expected.

#### `db/db_test.go`

**`db_test.go`**

```go
package db

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// setupMockDB initializes a mock database connection and returns the database instance and sqlmock instance.
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New() // Create a new mock database connection
	if err != nil {
		t.Fatalf("Error initializing mock database: %v", err)
	}
	return db, mock
}

// TestDBStore_AllMailboxes tests the AllMailboxes method.
func TestDBStore_AllMailboxes(t *testing.T) {
	tests := []struct {
		name            string
		expectedRows    []Mailbox
		expectedError   error
	}{
		{
			name: "Successful retrieval",
			expectedRows: []Mailbox{
				{ID: 1, MPIID: "mpi123", Token: "token123", CreatedAt: "2024-07-23 12:00:00"},
				{ID: 2, MPIID: "mpi456", Token: "token456", CreatedAt: "2024-07-23 13:00:00"},
			},
			expectedError: nil,
		},
		{
			name: "Database error",
			expectedRows: nil,
			expectedError: sql.ErrNoRows, // Example error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()
			defer func() {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("There were unfulfilled expectations: %s", err)
				}
			}()

			if tt.expectedError == nil {
				mock.ExpectQuery("SELECT id, mpi_id, token, created_at FROM mailboxes").
					WillReturnRows(sqlmock.NewRows([]string{"id", "mpi_id", "token", "created_at"}).
						AddRow(tt.expectedRows[0].ID, tt.expectedRows[0].MPIID, tt.expectedRows[0

].Token, tt.expectedRows[0].CreatedAt).
						AddRow(tt.expectedRows[1].ID, tt.expectedRows[1].MPIID, tt.expectedRows[1].Token, tt.expectedRows[1].CreatedAt))
			} else {
				mock.ExpectQuery("SELECT id, mpi_id, token, created_at FROM mailboxes").
					WillReturnError(tt.expectedError)
			}

			store := &DBStore{db: db}
			mailboxChan, err := store.AllMailboxes()
			if err != nil {
				if tt.expectedError == nil {
					t.Fatalf("Error calling AllMailboxes: %v", err)
				}
				return
			}

			var receivedMailboxes []Mailbox
			for mb := range mailboxChan {
				receivedMailboxes = append(receivedMailboxes, mb)
			}

			if len(receivedMailboxes) != len(tt.expectedRows) {
				t.Errorf("Expected %d mailboxes, got %d", len(tt.expectedRows), len(receivedMailboxes))
			}

			for i := range tt.expectedRows {
				if !reflect.DeepEqual(receivedMailboxes[i], tt.expectedRows[i]) {
					t.Errorf("Expected mailbox %v, got %v", tt.expectedRows[i], receivedMailboxes[i])
				}
			}
		})
	}
}

// TestDBStore_UsersForMailbox tests the UsersForMailbox method.
func TestDBStore_UsersForMailbox(t *testing.T) {
	tests := []struct {
		name            string
		mailboxID       int
		expectedRows    []User
		expectedError   error
	}{
		{
			name: "Successful retrieval",
			mailboxID: 1,
			expectedRows: []User{
				{ID: 101, MailboxID: 1, UserName: "user1", EmailAddress: "user1@example.com", CreatedAt: "2024-07-23 12:30:00"},
				{ID: 102, MailboxID: 1, UserName: "user2", EmailAddress: "user2@example.com", CreatedAt: "2024-07-23 12:45:00"},
			},
			expectedError: nil,
		},
		{
			name: "Database error",
			mailboxID: 1,
			expectedRows: nil,
			expectedError: sql.ErrNoRows, // Example error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()
			defer func() {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("There were unfulfilled expectations: %s", err)
				}
			}()

			if tt.expectedError == nil {
				mock.ExpectQuery("SELECT id, mailbox_id, user_name, email_address, created_at FROM users WHERE mailbox_id = ?").
					WithArgs(tt.mailboxID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "mailbox_id", "user_name", "email_address", "created_at"}).
						AddRow(tt.expectedRows[0].ID, tt.expectedRows[0].MailboxID, tt.expectedRows[0].UserName, tt.expectedRows[0].EmailAddress, tt.expectedRows[0].CreatedAt).
						AddRow(tt.expectedRows[1].ID, tt.expectedRows[1].MailboxID, tt.expectedRows[1].UserName, tt.expectedRows[1].EmailAddress, tt.expectedRows[1].CreatedAt))
			} else {
				mock.ExpectQuery("SELECT id, mailbox_id, user_name, email_address, created_at FROM users WHERE mailbox_id = ?").
					WithArgs(tt.mailboxID).
					WillReturnError(tt.expectedError)
			}

			store := &DBStore{db: db}
			userChan, err := store.UsersForMailbox(tt.mailboxID)
			if err != nil {
				if tt.expectedError == nil {
					t.Fatalf("Error calling UsersForMailbox: %v", err)
				}
				return
			}

			var receivedUsers []User
			for user := range userChan {
				receivedUsers = append(receivedUsers, user)
			}

			if len(receivedUsers) != len(tt.expectedRows) {
				t.Errorf("Expected %d users, got %d", len(tt.expectedRows), len(receivedUsers))
			}

			for i := range tt.expectedRows {
				if !reflect.DeepEqual(receivedUsers[i], tt.expectedRows[i]) {
					t.Errorf("Expected user %v, got %v", tt.expectedRows[i], receivedUsers[i])
				}
			}
		})
	}
}
```

1. **Test Cases**: Each test case covers different scenarios, including successful data retrieval and handling of database errors.

2. **Table-Driven Tests**: We use table-driven tests to simplify test case management and improve readability. Each table entry represents a different test scenario.

3. **Mocking**: We use `sqlmock` to simulate database queries and responses, allowing us to test our logic without relying on a real database.

4. **Error Handling**: Testing various error scenarios ensures that our code gracefully handles unexpected issues and provides meaningful error messages.


In this tutorial, we built a simple data pipeline in Go, separating the data access logic from the pipeline processing logic.

We used `sqlmock` for testing, which allows us to simulate database interactions and verify that our data access methods behave correctly. By organizing our code into well-defined modules and employing table-driven tests, we ensure that our data pipeline is robust and maintainable.

Feel free to expand on this tutorial by adding more features, handling additional edge cases, or integrating with other data sources.
