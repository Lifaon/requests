# Requests

A structure and its methods, to simplify MySQL requests.

## Installation

Simply use the `go get` command to install the package to your \$GOPATH:

```sh
go get github.com/mlantonn/requests
```

## Usage

Here is what a basic usage looks like:

```Golang
type Users []User
type User {
    ID    uint64
    Login string
}

// This method takes *sql.DB or *sql.Tx as an argument, and inserts all users
// inside the database, during or outside of a transaction
func (users Users) InsertIntoDB(sqlHandler interface{}) error {
    return requests.Request{
        SQLHandler: sqlHandler,
        Statement:  "INSERT INTO",
        Table:      "user",
        Condition:  "VALUES (?, ?)",
        ExecFunc:   insertUsers,
        Source:     users,
    }.ExecQuery()
}

// Request.ExecFunc for Users
func insertUsers(stmt *sql.Stmt, source interface{}) error {
    // Convert interface
    users, ok := source.(Users)
    if !ok {
        return fmt.Errorf("wrong type passed to insertUsers: %T (expected Users)", source)
    }
    // Insert all users
    for _, user := range users {
        _, err := stmt.Exec(user.ID, user.Login)
        if err != nil {
            return err
        }
    }
    return nil
}
```

### The Request structure

_Request_ is a structure which contains:

- An **SQLHandler** (can be _\*sql.DB_ or _\*sql.Tx_), used to make requests during or outside transactions.
- The query, taking form either:
  - As a whole: **Query**
  - In split parts: **Statement** (INSERT, UPDATE, etc), **Table** (targetted table), and **Condition** (WHERE, etc)
- A **ScanFunc** and its **Receiver**, to scan and store query results upon completion
- An **ExecFunc** andd its **Source**, to execute the query once or multiple times, with complex parameters

Here's the full structure:

```Golang
type Request struct {
    SQLHandler interface{} // might be *sql.Tx or *sql.DB
    Query      string      // full Query string
    Statement  string      // statement part of query
    Table      string      // targetted table of query
    Condition  string      // optionnal condition of query
    ScanFunc   scanFunc    // described below
    Receiver   interface{} // receiver for ScanFunc
    ExecFunc   execFunc    // described below
    Source     interface{} // source for ExecFunc
}
```

### The functions from Request structure

#### ScanFunc

```Golang
type scanFunc func(row interface{}, receiver interface{}) error
```

These functions are used to scan from *sql.Row or *sql.Rows (via already existing `ScanRow` function) and store the result in the given receiver, which should be a pointer to a structure, a slice of structures, or any other logical storage form. Here is an example of a `scanFunc`:<br />

```Golang
type User {
    ID    uint64
    Login string
}

func scanUser(row interface{}, ptr interface{}) error {
    // Unmarshal row
    var user User
    err := requests.ScanRow(row, &user.ID, &user.Login)
    if err != nil {
        return err
    }
    // Store result in value pointed by ptr
    switch r := ptr.(type) {
    case *[]User:
        *r = append(*r, user)
    case *User:
        *r = user
    default:
        return fmt.Errorf("wrong type passed to scanUser: %T (expect *[]User or *User)", r)
    }
    return nil
}
```

#### ExecFunc

```Golang
type execFunc func(stmt *sql.Stmt, source interface{}) error
```

These functions are used to execute a prepared statement multiple times, or to pass complex arguments. You have an example at the top of this README.<br />

### Methods

```Golang
// Prepares statements. Is always called in other methods
func (rq Request) PrepareStmt() (*sql.Stmt, error) {}

// Retrieve rows from query
func (rq Request) GetRows(args ...interface{}) (*sql.Rows, error) {}

// Retrieve rows from query + Scan rows in receiver
func (rq Request) GetRowsAndScan(args ...interface{}) error {}

// Retrieve first row from query
func (rq Request) GetOneRow(args ...interface{}) (*sql.Row, error) {}

// Retrieve first row from query + Scan row in receiver
func (rq Request) GetOneRowAndScan(args ...interface{}) error {}

// Execute query, via given function and source parameters if present
func (rq Request) ExecQuery() error {}
```
