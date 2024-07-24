package db

type Mailbox struct {
		ID        int
		MPIID     string
		Token     string
		CreatedAt string
}

type User struct {
		ID           int
		MailboxID    int
		UserName     string
		EmailAddress string
		CreatedAt    string
}

type Store interface {
		AllMailboxes() ([]Mailbox, error)
		UsersForMailbox(mailboxID int) ([]User, error)
}
