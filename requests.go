package requests

import (
	"database/sql"
	"fmt"
)

type (
	// Request is used to prepare (and optionnaly make) queries via SQLHandler
	// (which can be *sql.DB or *sql.Tx), during or outside transactions.
	Request struct {
		SQLHandler interface{} // might be *sql.Tx or *sql.DB
		Query      Query       // Query structure
		ExecFunc   execFunc    // described below
		Arg        interface{} // for ExecFunc or scan functions
	}

	// Query is a structure used to concatenate queries
	Query struct {
		Query     string // full Query string
		Statement string // statement part of query
		Table     string // targetted table of query
		Condition string // optionnal condition of query
	}

	// Exec function which takes *sql.Stmt as a first argument, and executes the
	// prepared query with the given arguments
	execFunc func(stmt *sql.Stmt, source interface{}) error
)

// String checks creates a Query string from its other parameters if its Query
// parameter is empty
func (q Query) String() string {
	if q.Query == "" {
		return q.Statement + " " + q.Table + " " + q.Condition
	}
	return q.Query
}

// PrepareStmt prepares rq.Query via rq.SQLHandler. If rq.Query is empty, it
// will concatenate the query with rq.Stmt, rq.DBTable, and rq.Condition
func (rq Request) PrepareStmt() (*sql.Stmt, error) {
	switch r := rq.SQLHandler.(type) {
	case *sql.DB:
		return r.Prepare(rq.Query.String())
	case *sql.Tx:
		return r.Prepare(rq.Query.String())
	default:
		return nil, fmt.Errorf("wrong type passed to PrepareStatementQuery %T (expected *sql.DB or *sql.Tx)", r)
	}

}

// GetRows makes a prepared query and returns the resulted rows. This function
// can be used during and outside of transactions.
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

// GetRowsAndScan retrieves rows from given query, then calls the passed
// ScanFunc for each row to store results directly into the Receiver (usually
// pointing to a slice of structures)
func (rq Request) GetRowsAndScan(args ...interface{}) error {

	// Retrieve rows
	rows, err := rq.GetRows(args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Scan rows into receiver
	return scanRows(rows, rq.Arg)
}

// GetOneRow makes a prepared query and returns the resulted row. This function
// can be used during and outside of transactions.
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

// GetOneRowAndScan retrieves the first row from given query, then calls the
// passed ScanFunc to store results directly into the Receiver (usually pointing
// to a structure)
func (rq Request) GetOneRowAndScan(args ...interface{}) error {

	// Retrieve row
	row, err := rq.GetOneRow(args...)
	if err != nil {
		return err
	}

	// Scan row into receiver
	return scanOneRow(row, rq.Arg)
}

// ExecQuery prepares a query which does not return a row, then calls the given
// ExecFunc with the passed Arg. This function can be used during and outside
// of transactions.
func (rq Request) ExecQuery(args ...interface{}) error {

	// Prepare statement
	stmt, err := rq.PrepareStmt()
	if err != nil {
		return err
	}
	defer stmt.Close()

	// If an execFunc was given, execute it
	if rq.ExecFunc != nil {
		return rq.ExecFunc(stmt, rq.Arg)
	}

	// Otherwise, execute statement with given arguments
	_, err = stmt.Exec(args...)
	return err
}
