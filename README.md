# Mailboxes

This Go program manages mailboxes and users by interacting with a database using the `database/sql` package. It utilizes `go-sqlmock` for mocking database interactions during testing.

## Implementation

### Components

The program consists of the following main components:

1. **Mailbox Struct**:
	 - Represents a mailbox with attributes such as ID, MPIID, Token, and CreatedAt.

2. **User Struct**:
	 - Represents a user associated with a mailbox, containing attributes like ID, MailboxID, UserName, EmailAddress, and CreatedAt.

3. **Database Interaction**:
	 - Uses `database/sql` to connect to a SQLite database (`github.com/mattn/go-sqlite3` driver).
	 - Implements functions to retrieve mailboxes (`RetrieveMailboxes`) and users (`RetrieveUsersForMailbox`) from the database using SQL queries.

4. **Processing Logic**:
	 - **processUser(user User)**:
		 - A fictional function that logs processing information about each user.
	 - **Pipeline(db \*sql.DB)**:
		 - Orchestrates the processing of mailboxes and users concurrently using goroutines and `sync.WaitGroup`.

5. **Main Function**:
	 - Initializes configuration using `github.com/spf13/viper` from a `config.yaml` file.
	 - Sets up database connection based on configuration.
	 - Calls `Pipeline(db)` to initiate processing of mailboxes and users.

### Concurrency and Channel Usage

- **RetrieveMailboxes**:
	- Retrieves mailboxes from the database and sends them over a channel (`mailboxChannel`), *dynamically sized based on the count of retrieved records*.

- **RetrieveUsersForMailbox**:
	- Retrieves users associated with a specific mailbox ID and sends them over a channel (`userChannel`).

- **Pipeline Function**:
	- Concurrently processes each retrieved mailbox, spawns a goroutine for each mailbox to process its associated users

## Setup Locally

To set up and run the Mailboxes program locally, follow these steps:

### Prerequisites

- Go programming language installed (version 1.16 or higher recommended).
- SQLite3 installed (if using SQLite as the database).

### Steps

1. **Clone the Repository**

	 ```bash
	 git clone <repository-url>
	 cd <repository-directory>
	 ```

2. **Install Dependencies**

	 Use Go modules to install dependencies:

	 ```bash
	 go mod tidy
	 ```

3. **Create Configuration File**

	 Create a `config.yaml` file in the root directory with the following content:

	 ```yaml
	 database:
		 driver: sqlite3
		 path: ./test.db
	 ```

	 Adjust `path` as necessary to point to the SQLite database file.

4. **Initialize Database**

	 Initialize the SQLite database with necessary tables and sample data.
  ```bash
    sqlite3 test.db < schema.sql
  ```


5. **Build and Run**

	 Build and run the program using the following commands:

	 ```bash
	 go build -o mailboxes .
	 ./mailboxes
	 ```

	 This will compile the Go code into an executable `mailboxes` and execute it, which will connect to the database, retrieve mailboxes, and process users accordingly.

## Running Tests

To run tests for the Mailboxes program, follow these steps:

1. **Ensure Dependencies**

	 Ensure that all dependencies, including testing libraries (`go-sqlmock`), are installed (already handled if `go mod tidy` was run).

2. **Run Tests**

	 Execute the tests using the `go test` command:

	 ```bash
	 go test ./...
	 ```

	 This command runs all tests in the current directory and subdirectories (`./...`). The tests will verify various aspects of the program, including database query execution, data retrieval, and processing logic.
