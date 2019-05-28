[![Build Status](https://travis-ci.org/mlantonn/requests.svg?branch=master)](https://travis-ci.org/mlantonn/requests) [![codecov](https://codecov.io/gh/mlantonn/requests/branch/master/graph/badge.svg)](https://codecov.io/gh/mlantonn/requests) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/mlantonn/requests) [![Go Report Card](https://goreportcard.com/badge/github.com/mlantonn/requests)](https://goreportcard.com/report/github.com/mlantonn/requests)

# Requests

A structure and its methods, to simplify SQL requests.

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
// PrepareStmt prepares rq.Query via rq.Handler. If rq.Query is empty, it
// will concatenate the query with its subparts (Statement, Table, ...)
func (rq Request) PrepareStmt() (*sql.Stmt, error) {}

// InsertStructs inserts a slice of structures into the given table. Can insert
// sub structures, if they have the tag `req:"include"`
func (rq Request) InsertStructs(slice interface{}) error {}

// InsertOneStruct inserts one structure into the given table. Can insert sub
// structures, if they have the tag `req:"include"`
func (rq Request) InsertOneStruct(structure interface{}) error {}

// GetRows prepares and makes a query, and returns the resulted rows
func (rq Request) GetRows(args ...interface{}) (*sql.Rows, error) {}

// GetOneRow prepares and makes a query, and returns the resulted row
func (rq Request) GetOneRow(args ...interface{}) (*sql.Row, error) {}

// GetIntoStructs prepares and makes a query, retrieves Rows, and scan them into
// a pointer to a slice of structures. Should be used for queries selecting
// multiple fields from multiple rows
func (rq Request) GetIntoStructs(args ...interface{}) error {}

// GetIntoSlice prepares and makes a query, retrieves Rows, and scan them into
// a pointer to a slice of any type. Should be used for queries selecting one
// field from multiple rows
func (rq Request) GetIntoSlice(args ...interface{}) error {}

// GetOneField prepares and makes a query, retrieves a Row, and scan it into
// the passed pointer. Should be used for queries selecting one field from one
// row
func (rq Request) GetOneField(args ...interface{}) error {}

// GetFields prepares and makes a query, retrieves a Row, and scan it into
// a slice of any pointers. Should be used for queries selecting multiple fields
// from a single row
func (rq Request) GetFields(args ...interface{}) error {}

// GetIntoOneStruct prepares and makes a query, retrieves a Row, and scan it
// into a pointer to one struct. Should be used for queries selecting multiple
// fields from a single row (if this query is unique, consider using GetFields
// instead)
func (rq Request) GetIntoOneStruct(args ...interface{}) error {}

// ExecQuery prepares and makes a query which does not need to return any row
func (rq Request) ExecQuery(args ...interface{}) (sql.Result, error) {}

```
