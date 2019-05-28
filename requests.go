package requests

import (
	"database/sql"
	"errors"
)

type (
	// Handler implements Prepare (from either *sql.DB or *sql.Tx)
	Handler interface {
		Prepare(query string) (*sql.Stmt, error)
	}

	// Request is used to prepare (and optionnaly make) queries via Handler
	// (which can be *sql.DB or *sql.Tx), during or outside transactions.
	Request struct {
		Handler                 // might be *sql.Tx or *sql.DB
		QueryStruct             // Query structure
		Arg         interface{} // for scan functions
	}

	// QueryStruct is a structure used to concatenate queries
	QueryStruct struct {
		Query     string // full Query string
		Statement string // statement part of query
		Table     string // targeted table of query
		Set       string // optionnal set parameter of query
		Condition string // optionnal condition of query
	}
)

// ErrNoArg is for functions that should scan rows but have no passed argument
var ErrNoArg = errors.New("no passed argument, can not scan")

// FromHandler returns an initialized Request with given Handler (*Tx or *DB)
func FromHandler(handler Handler) Request {
	return Request{Handler: handler}
}

// String returns a Query string from QueryStruct. If the Query field is not
// empty, it is returned, else a concatenation of other fields is returned
func (q QueryStruct) String() string {
	if q.Query == "" {
		s := q.Statement + " " + q.Table
		if q.Set != "" {
			s += " " + q.Set
		}
		if q.Condition != "" {
			s += " " + q.Condition
		}
		return s
	}
	return q.Query
}

// PrepareStmt prepares rq.Query via rq.Handler. If rq.Query is empty, it
// will concatenate the query with its subparts (Statement, Table, ...).
// Always called in other methods
func (rq Request) PrepareStmt() (*sql.Stmt, error) {
	return rq.Handler.Prepare(rq.QueryStruct.String())
}

// GetRows prepares and makes a query, and returns the resulted rows
func (rq Request) GetRows(args ...interface{}) (*sql.Rows, error) {

	// Prepare statement
	stmt, err := rq.PrepareStmt()
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Retrieve rows
	return stmt.Query(args...)
}

// GetOneRow prepares and makes a query, and returns the resulted row
func (rq Request) GetOneRow(args ...interface{}) (*sql.Row, error) {

	// Prepare statement
	stmt, err := rq.PrepareStmt()
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Retrieve row
	return stmt.QueryRow(args...), nil
}

// ExecQuery prepares and makes a query which does not need to return any row
func (rq Request) ExecQuery(args ...interface{}) (sql.Result, error) {

	// Prepare statement
	stmt, err := rq.PrepareStmt()
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Execute statement with given arguments
	return stmt.Exec(args...)
}
