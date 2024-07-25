package main

import (
	"log"
	"testing"
	"time"
	"sync"

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
func (m *MockStore) AllMailboxes() (<-chan db.Mailbox, error) {
	m.CountCalls++
	mailboxChan := make(chan db.Mailbox)

	go func() {
		defer close(mailboxChan)
		for _, mb := range m.Mailboxes {
			mailboxChan <- mb
		}
	}()

	return mailboxChan, m.Err
}

// UsersForMailbox mocks retrieving users for a mailbox ID
func (m *MockStore) UsersForMailbox(mailboxID int) (<-chan db.User, error) {
	m.CountCalls++
	userChan := make(chan db.User)

	go func() {
		defer close(userChan)
		users, ok := m.Users[mailboxID]
		if !ok {
			return
		}
		for _, user := range users {
			userChan <- user
		}
	}()

	return userChan, m.Err
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
	mailboxChan := RetrieveMailboxes(mockStore)

	// Verify the received mailboxes
	var receivedMailboxes []db.Mailbox
	for mb := range mailboxChan {
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
	userChan := RetrieveUsersForMailbox(mockStore, 1)

	// Verify the received users
	var receivedUsers []db.User
	for user := range userChan {
		receivedUsers = append(receivedUsers, user)
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

	// Set up a wait group to synchronize goroutines in Pipeline
	var wg sync.WaitGroup

	// Mock function to process a user
	processUser := func(user db.User) {
		log.Printf("Processing user: User Name - %s, Mailbox Token - %s", user.UserName, "<fake_token>")
	}

	// Mock pipeline function
	pipeline := func(store db.Store) {
		mailboxChan := RetrieveMailboxes(store)

		for mb := range mailboxChan {
			log.Printf("Processing %d mailbox", mb.ID)

			userChan := RetrieveUsersForMailbox(store, mb.ID)
			wg.Add(1)

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
	}

	// Call Pipeline
	go pipeline(mockStore)

	// Allow some time for all goroutines to finish
	time.Sleep(1 * time.Second)

	// Wait for all goroutines to finish
	wg.Wait()
}
