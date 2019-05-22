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
    ID    uint64 `db:"id"`
    Login string `db:"login"`
}

func (users Users) Insert(handler requests.Handler) error {
    rq := requests.FromHandler(handler)
    rq.Table = "user"
    return rq.InsertStructs(users)
}
```

### The Request structure

_Request_ is a structure which contains:

- The **Handler** (can be _\*sql.DB_ or _\*sql.Tx_), used to make requests during or outside transactions.
- The query, taking form either:
  - As a whole: **Query**
  - In split parts: **Statement**, **Table**, **Set** and **Condition**
- An optional **Arg** (eg: to store query results)

Here's the full architecture:

```Golang
type (
    // Handler implements Prepare (from either *sql.DB or *sql.Tx)
    Handler interface {
        Prepare(query string) (*sql.Stmt, error)
    }

    // Request is used to prepare (and optionnaly make) queries via Handler
    // (which can be *sql.DB or *sql.Tx), during or outside transactions.
    Request struct {
        Handler             // might be *sql.Tx or *sql.DB
        query               // Query structure
        Arg     interface{} // for scan functions
    }

    // query is a structure used to concatenate queries
    query struct {
        Query     string // full Query string
        Statement string // statement part of query
        Table     string // targetted table of query
        Set       string // optionnal set parameter of query
        Condition string // optionnal condition of query
    }
)
```

### Methods

```Golang
// Prepares statements. Is always called in other methods
func (rq Request) PrepareStmt() (*sql.Stmt, error) {}

// Inserts a slice of structures into the given table
func (rq Request) InsertStructs(slice interface{}) error {}

// Inserts a single structure into the given table
func (rq Request) InsertOneStruct(structure interface{}) error {}

// Prepare and make query, return rows
func (rq Request) GetRows(args ...interface{}) (*sql.Rows, error) {}

// Prepare and make query, scan rows into slice of structures
func (rq Request) GetIntoStructs(args ...interface{}) error {}

// Prepare and make query, scan rows into basic slice
func (rq Request) GetIntoSlice(args ...interface{}) error {}

// Prepare and make query, return row
func (rq Request) GetOneRow(args ...interface{}) (*sql.Row, error) {}

// Prepare and make query, scan row into interface{}
func (rq Request) GetOneField(args ...interface{}) error {}

// Prepare and make query, scan rows into slice of ptr
func (rq Request) GetFields(args ...interface{}) error {}

// Prepare and make query, scan rows into structure
func (rq Request) GetIntoOneStruct(args ...interface{}) error {}

// Prepare and make query which doesn't retrieve rows
func (rq Request) ExecQuery(args ...interface{}) (sql.Result, error) {}

```
