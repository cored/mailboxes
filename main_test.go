package main

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRetrieveMailboxes(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	defer func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	}()

	expectedMailboxes := []Mailbox{
		{ID: 1, MPIID: "mpi123", Token: "token123", CreatedAt: "2024-07-23 12:00:00"},
		{ID: 2, MPIID: "mpi456", Token: "token456", CreatedAt: "2024-07-23 13:00:00"},
	}

	// Expecting query to retrieve mailboxes
	mock.ExpectQuery("SELECT id, mpi_id, token, created_at FROM mailboxes").
		WillReturnRows(sqlmock.NewRows([]string{"id", "mpi_id", "token", "created_at"}).
			AddRow(expectedMailboxes[0].ID, expectedMailboxes[0].MPIID, expectedMailboxes[0].Token, expectedMailboxes[0].CreatedAt).
			AddRow(expectedMailboxes[1].ID, expectedMailboxes[1].MPIID, expectedMailboxes[1].Token, expectedMailboxes[1].CreatedAt))

	// Call RetrieveMailboxes
	mailboxes := RetrieveMailboxes(db)

	// Verify the received mailboxes
	var receivedMailboxes []Mailbox
	for mb := range mailboxes {
		receivedMailboxes = append(receivedMailboxes, mb)
	}

	if len(receivedMailboxes) != len(expectedMailboxes) {
		t.Errorf("Expected %d mailboxes, got %d", len(expectedMailboxes), len(receivedMailboxes))
	}

	for i := range expectedMailboxes {
		if receivedMailboxes[i] != expectedMailboxes[i] {
			t.Errorf("Expected mailbox %v, got %v", expectedMailboxes[i], receivedMailboxes[i])
		}
	}
}

func TestRetrieveUsersForMailbox(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	defer func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	}()

	mailboxID := 1
	expectedUsers := []User{
		{ID: 101, MailboxID: mailboxID, UserName: "user1", EmailAddress: "user1@example.com", CreatedAt: "2024-07-23 12:30:00"},
		{ID: 102, MailboxID: mailboxID, UserName: "user2", EmailAddress: "user2@example.com", CreatedAt: "2024-07-23 12:45:00"},
	}

	// Expecting query to retrieve users for mailboxID
	mock.ExpectQuery("SELECT id, mailbox_id, user_name, email_address, created_at FROM users WHERE mailbox_id = ?").
		WithArgs(mailboxID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "mailbox_id", "user_name", "email_address", "created_at"}).
			AddRow(expectedUsers[0].ID, expectedUsers[0].MailboxID, expectedUsers[0].UserName, expectedUsers[0].EmailAddress, expectedUsers[0].CreatedAt).
			AddRow(expectedUsers[1].ID, expectedUsers[1].MailboxID, expectedUsers[1].UserName, expectedUsers[1].EmailAddress, expectedUsers[1].CreatedAt))

	// Call RetrieveUsersForMailbox
	users := RetrieveUsersForMailbox(db, mailboxID)

	// Verify the received users
	var receivedUsers []User
	for u := range users {
		receivedUsers = append(receivedUsers, u)
	}

	if len(receivedUsers) != len(expectedUsers) {
		t.Errorf("Expected %d users, got %d", len(expectedUsers), len(receivedUsers))
	}

	for i := range expectedUsers {
		if receivedUsers[i] != expectedUsers[i] {
			t.Errorf("Expected user %v, got %v", expectedUsers[i], receivedUsers[i])
		}
	}
}

func TestPipeline(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	defer func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	}()

	// Mocking RetrieveMailboxes
	expectedMailboxes := []Mailbox{
		{ID: 1, MPIID: "mpi123", Token: "token123", CreatedAt: "2024-07-23 12:00:00"},
		{ID: 2, MPIID: "mpi456", Token: "token456", CreatedAt: "2024-07-23 13:00:00"},
	}

	mock.ExpectQuery("SELECT id, mpi_id, token, created_at FROM mailboxes").
		WillReturnRows(sqlmock.NewRows([]string{"id", "mpi_id", "token", "created_at"}).
			AddRow(expectedMailboxes[0].ID, expectedMailboxes[0].MPIID, expectedMailboxes[0].Token, expectedMailboxes[0].CreatedAt).
			AddRow(expectedMailboxes[1].ID, expectedMailboxes[1].MPIID, expectedMailboxes[1].Token, expectedMailboxes[1].CreatedAt))

	// Mocking RetrieveUsersForMailbox for each mailbox
	for _, mb := range expectedMailboxes {
		mock.ExpectQuery("SELECT id, mailbox_id, user_name, email_address, created_at FROM users WHERE mailbox_id = ?").
			WithArgs(mb.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "mailbox_id", "user_name", "email_address", "created_at"}))
	}

	// Call Pipeline
	Pipeline(db)

	// Verify that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

// Mock database setup for testing
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing mock database: %v", err)
	}
	return db, mock
}
