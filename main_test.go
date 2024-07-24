package main

import (
	"testing"
	"mailboxes/db"

)

// MockStore is a fake implementation of Store for testing purposes
type MockStore struct {
	Mailboxes  []db.Mailbox
	Users      map[int][]db.User
	Err        error
	CountCalls int
}

// AllMailboxes mocks retrieving all mailboxes
func (m *MockStore) AllMailboxes() ([]db.Mailbox, error) {
	m.CountCalls++
	return m.Mailboxes, m.Err
}

// UsersForMailbox mocks retrieving users for a mailbox ID
func (m *MockStore) UsersForMailbox(mailboxID int) ([]db.User, error) {
	m.CountCalls++
	return m.Users[mailboxID], m.Err
}

// TestRetrieveMailboxes tests RetrieveMailboxes function using MockStore
func TestRetrieveMailboxes(t *testing.T) {
	mockStore := &MockStore{
		Mailboxes: []db.Mailbox{
			{ID: 1, MPIID: "mpi123", Token: "token123", CreatedAt: "2024-07-23 12:00:00"},
			{ID: 2, MPIID: "mpi456", Token: "token456", CreatedAt: "2024-07-23 13:00:00"},
		},
	}

	// Call RetrieveMailboxes
	mailboxes := RetrieveMailboxes(mockStore)

	// Verify the received mailboxes
	var receivedMailboxes []db.Mailbox
	for mb := range mailboxes {
		receivedMailboxes = append(receivedMailboxes, mb)
	}

	expectedMailboxes := mockStore.Mailboxes
	if len(receivedMailboxes) != len(expectedMailboxes) {
		t.Errorf("Expected %d mailboxes, got %d", len(expectedMailboxes), len(receivedMailboxes))
	}

	for i := range expectedMailboxes {
		if receivedMailboxes[i] != expectedMailboxes[i] {
			t.Errorf("Expected mailbox %v, got %v", expectedMailboxes[i], receivedMailboxes[i])
		}
	}
}

// TestRetrieveUsersForMailbox tests RetrieveUsersForMailbox function using MockStore
func TestRetrieveUsersForMailbox(t *testing.T) {
	mockStore := &MockStore{
		Users: map[int][]db.User{
			1: {
				{ID: 101, MailboxID: 1, UserName: "user1", EmailAddress: "user1@example.com", CreatedAt: "2024-07-23 12:30:00"},
				{ID: 102, MailboxID: 1, UserName: "user2", EmailAddress: "user2@example.com", CreatedAt: "2024-07-23 12:45:00"},
			},
		},
	}

	// Call RetrieveUsersForMailbox
	users := RetrieveUsersForMailbox(mockStore, 1)

	// Verify the received users
	var receivedUsers []db.User
	for u := range users {
		receivedUsers = append(receivedUsers, u)
	}

	expectedUsers := mockStore.Users[1]
	if len(receivedUsers) != len(expectedUsers) {
		t.Errorf("Expected %d users, got %d", len(expectedUsers), len(receivedUsers))
	}

	for i := range expectedUsers {
		if receivedUsers[i] != expectedUsers[i] {
			t.Errorf("Expected user %v, got %v", expectedUsers[i], receivedUsers[i])
		}
	}
}

// TestPipeline tests the entire pipeline using MockStore
func TestPipeline(t *testing.T) {
	mockStore := &MockStore{
		Mailboxes: []db.Mailbox{
			{ID: 1, MPIID: "mpi123", Token: "token123", CreatedAt: "2024-07-23 12:00:00"},
			{ID: 2, MPIID: "mpi456", Token: "token456", CreatedAt: "2024-07-23 13:00:00"},
		},
		Users: map[int][]db.User{
			1: {
				{ID: 101, MailboxID: 1, UserName: "user1", EmailAddress: "user1@example.com", CreatedAt: "2024-07-23 12:30:00"},
				{ID: 102, MailboxID: 1, UserName: "user2", EmailAddress: "user2@example.com", CreatedAt: "2024-07-23 12:45:00"},
			},
			2: {
				{ID: 201, MailboxID: 2, UserName: "user3", EmailAddress: "user3@example.com", CreatedAt: "2024-07-23 13:30:00"},
				{ID: 202, MailboxID: 2, UserName: "user4", EmailAddress: "user4@example.com", CreatedAt: "2024-07-23 13:45:00"},
			},
		},
	}

	// Call Pipeline
	Pipeline(mockStore)
}
