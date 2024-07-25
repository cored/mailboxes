package db

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDBStore_AllMailboxes(t *testing.T) {
	tests := []struct {
		name           string
		expectedMailboxes []Mailbox
		mockRows       *sqlmock.Rows
		expectedError  error
	}{
		{
			name: "Success with multiple mailboxes",
			expectedMailboxes: []Mailbox{
				{ID: 1, MPIID: "mpi123", Token: "token123", CreatedAt: "2024-07-23 12:00:00"},
				{ID: 2, MPIID: "mpi456", Token: "token456", CreatedAt: "2024-07-23 13:00:00"},
			},
			mockRows: sqlmock.NewRows([]string{"id", "mpi_id", "token", "created_at"}).
			AddRow(1, "mpi123", "token123", "2024-07-23 12:00:00").
			AddRow(2, "mpi456", "token456", "2024-07-23 13:00:00"),
			expectedError: nil,
		},
		{
			name: "No mailboxes",
			expectedMailboxes: []Mailbox{},
			mockRows: sqlmock.NewRows([]string{"id", "mpi_id", "token", "created_at"}),
			expectedError: nil,
		},
		{
			name: "Error retrieving mailboxes",
			expectedMailboxes: nil,
			mockRows: sqlmock.NewRows([]string{}),
			expectedError: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			// Setup mock expectations
			if tt.expectedError != nil {
				mock.ExpectQuery("SELECT id, mpi_id, token, created_at FROM mailboxes").WillReturnError(tt.expectedError)
			} else {
				mock.ExpectQuery("SELECT id, mpi_id, token, created_at FROM mailboxes").WillReturnRows(tt.mockRows)
			}

			store := &DBStore{db: db}

			// Call AllMailboxes method
			mailboxChan, err := store.AllMailboxes()
			if err != nil {
				if tt.expectedError == nil {
					t.Fatalf("Error calling AllMailboxes: %v", err)
				}
				if err != tt.expectedError {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			// Verify the received mailboxes
			var receivedMailboxes []Mailbox
			for mb := range mailboxChan {
				receivedMailboxes = append(receivedMailboxes, mb)
			}

			if len(receivedMailboxes) != len(tt.expectedMailboxes) {
				t.Errorf("Expected %d mailboxes, got %d", len(tt.expectedMailboxes), len(receivedMailboxes))
			}

			for i := range tt.expectedMailboxes {
				if !reflect.DeepEqual(receivedMailboxes[i], tt.expectedMailboxes[i]) {
					t.Errorf("Expected mailbox %v, got %v", tt.expectedMailboxes[i], receivedMailboxes[i])
				}
			}
		})
	}
}

func TestDBStore_UsersForMailbox(t *testing.T) {
	tests := []struct {
		name           string
		mailboxID      int
		expectedUsers  []User
		mockRows       *sqlmock.Rows
		expectedError  error
	}{
		{
			name:      "Success with multiple users",
			mailboxID: 1,
			expectedUsers: []User{
				{ID: 101, MailboxID: 1, UserName: "user1", EmailAddress: "user1@example.com", CreatedAt: "2024-07-23 12:30:00"},
				{ID: 102, MailboxID: 1, UserName: "user2", EmailAddress: "user2@example.com", CreatedAt: "2024-07-23 12:45:00"},
			},
			mockRows: sqlmock.NewRows([]string{"id", "mailbox_id", "user_name", "email_address", "created_at"}).
			AddRow(101, 1, "user1", "user1@example.com", "2024-07-23 12:30:00").
			AddRow(102, 1, "user2", "user2@example.com", "2024-07-23 12:45:00"),
			expectedError: nil,
		},
		{
			name:      "No users",
			mailboxID: 1,
			expectedUsers: []User{},
			mockRows: sqlmock.NewRows([]string{"id", "mailbox_id", "user_name", "email_address", "created_at"}),
			expectedError: nil,
		},
		{
			name:      "Error retrieving users",
			mailboxID: 1,
			expectedUsers: nil,
			mockRows: sqlmock.NewRows([]string{}),
			expectedError: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			// Setup mock expectations
			if tt.expectedError != nil {
				mock.ExpectQuery("SELECT id, mailbox_id, user_name, email_address, created_at FROM users WHERE mailbox_id = ?").
				WithArgs(tt.mailboxID).
				WillReturnError(tt.expectedError)
			} else {
				mock.ExpectQuery("SELECT id, mailbox_id, user_name, email_address, created_at FROM users WHERE mailbox_id = ?").
				WithArgs(tt.mailboxID).
				WillReturnRows(tt.mockRows)
			}

			store := &DBStore{db: db}

			// Call UsersForMailbox method
			userChan, err := store.UsersForMailbox(tt.mailboxID)
			if err != nil {
				if tt.expectedError == nil {
					t.Fatalf("Error calling UsersForMailbox: %v", err)
				}
				if err != tt.expectedError {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			// Verify the received users
			var receivedUsers []User
			for user := range userChan {
				receivedUsers = append(receivedUsers, user)
			}

			if len(receivedUsers) != len(tt.expectedUsers) {
				t.Errorf("Expected %d users, got %d", len(tt.expectedUsers), len(receivedUsers))
			}

			for i := range tt.expectedUsers {
				if !reflect.DeepEqual(receivedUsers[i], tt.expectedUsers[i]) {
					t.Errorf("Expected user %v, got %v", tt.expectedUsers[i], receivedUsers[i])
				}
			}
		})
	}
}

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New() // Create a new mock database connection
	if err != nil {
		t.Fatalf("Error initializing mock database: %v", err)
	}
	return db, mock
}
