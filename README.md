# Mailbox Processing System

## Implementation Overview

### 1. Database Connection and Configuration

- **Configuration Management**:
	- The system initializes a database connection using configuration loaded from a `config.yaml` file. Configuration management is handled using the `github.com/spf13/viper` package.
	- It establishes a connection to an SQLite database using the database driver (`dbDriver`) and path (`dbPath`) specified in the configuration file.

### 2. DBStore Implementation

- **DBStore Struct**:
	- The `DBStore` struct implements the `db.Store` interface, providing methods to interact with the database.
	- Key methods include:
		- `AllMailboxes()`: Retrieves all mailboxes from the database and returns a channel (`<-chan db.Mailbox`) that streams each mailbox as it's fetched.
		- `UsersForMailbox(mailboxID int)`: Retrieves users associated with a specific mailbox ID and returns a channel (`<-chan db.User`) that streams each user record.

### 3. Pipeline Function (`Pipeline`)

- **Functionality**:
	- The `Pipeline` function coordinates the process of retrieving mailboxes and their associated users.
	- It starts by fetching mailboxes using `store.AllMailboxes()`, which returns a channel of `Mailbox` objects.
	- For each retrieved mailbox, it concurrently retrieves users using `store.UsersForMailbox(mb.ID)` and processes each user in a separate goroutine.
	- A `sync.WaitGroup` is used to ensure all user processing goroutines complete before the function finishes.

## Trade-offs

### 1. Concurrency vs. Resource Consumption

- **Pros**:
	- Goroutines (`go` keyword) enable concurrent fetching and processing of data, enhancing performance and throughput, particularly for I/O-bound operations.
- **Cons**:
	- Managing multiple goroutines and synchronizing them using `sync.WaitGroup` adds complexity and requires careful handling to avoid race conditions and resource contention.

### 2. Memory Usage

- **Pros**:
	- Channels (`<-chan`) allow streaming of data, reducing memory overhead by processing items as they are retrieved from the database.
- **Cons**:
	- Open channels and goroutines can increase memory usage, especially with large datasets. Proper management and cleanup of resources are essential.

### 3. Error Handling and Debugging

- **Pros**:
	- The implementation includes comprehensive error handling (`err` checks) to address potential database connection issues or query failures.
- **Cons**:
	- Debugging concurrent code and managing errors across multiple goroutines can be challenging, necessitating careful logging and error handling strategies.

### 4. Scalability

- **Pros**:
	- The use of concurrency and data streaming via channels supports scalable data retrieval and processing, suitable for large datasets.
- **Cons**:
	- Horizontal scaling (across multiple servers or nodes) requires additional considerations for managing database connections and synchronization in distributed environments.

## Setup and Usage

### 1. Setting Up the Database Locally

1. **Create Database Schema**:
	 - Use the provided SQL script to set up the database schema and sample data. Save the following script as `schema.sql`:

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

		 -- Insert sample data into mailboxes table
		 INSERT INTO mailboxes (id, mpi_id, token, created_at)
		 VALUES
				 (1, 'mpi123', 'token123', '2024-07-23 12:00:00'),
				 (2, 'mpi456', 'token456', '2024-07-23 13:00:00');

		 -- Insert sample data into users table
		 INSERT INTO users (id, mailbox_id, user_name, email_address, created_at)
		 VALUES
				 (101, 1, 'user1', 'user1@example.com', '2024-07-23 12:30:00'),
				 (102, 1, 'user2', 'user2@example.com', '2024-07-23 12:45:00'),
				 (201, 2, 'user3', 'user3@example.com', '2024-07-23 13:15:00');
		 ```

2. **Execute the Script**:
	 - Run the script `bin/dbsetup` to setup the database and add test data:
		 ```sh
		./bin/dbsetup
		 ```

### 2. Running the Program

1. **Build the Application**:
	 - Navigate to the project directory and execute `/bin/dev`	to build and run
	 the application:
		 ```sh
		 ./bin/dev
		 ```

### 3. Running the Tests

1. **Run Unit Tests**:
	 - Ensure you have Go installed and the `go test` command available.
	 - Run the tests using:
		 ```sh
		 go test -v
		 ```

2. **Test Output**:
	 - Review test results in the console output for verification of functionality and correctness.

## Configuration

- **Configuration File**:
	- Create a `config.yaml` file in the root directory with the following structure:
		```yaml
		database:
			driver: sqlite3
			path: path_to_your_database.db
		```

- **Adjust the `path`** according to your local database file location.

## Additional Information

- **Dependencies**:
	- The project depends on the `github.com/spf13/viper` package for configuration management. Ensure it is included in your `go.mod` file.

- **Logging**:
	- The application uses standard logging to provide runtime information and debugging output.
