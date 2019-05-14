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
		Arg        interface{} // for scan functions
	}

	// Query is a structure used to concatenate queries
	Query struct {
		Query     string // full Query string
		Statement string // statement part of query
		Table     string // targetted table of query
		Set       string // optionnal set parameter of query
		Condition string // optionnal condition of query
	}
)

// FromHandler returns an initialized Request with given SQL Handler (Tx or DB)
func FromHandler(handler interface{}) Request {
	return Request{SQLHandler: handler}
}

// String checks creates a Query string from its other parameters if its Query
// parameter is empty
func (q Query) String() string {
	if q.Query == "" {
		return q.Statement + " " + q.Table + " " + q.Set + " " + q.Condition
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

// GetIntoStructs retrieves rows from given query, then calls the passed
// ScanFunc for each row to store results directly into the Receiver (usually
// pointing to a slice of structures)
func (rq Request) GetIntoStructs(args ...interface{}) error {

	// Retrieve rows
	rows, err := rq.GetRows(args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Scan rows into receiver
	return scanToSliceOfStruct(rows, rq.Arg)
}

// GetIntoSlice retrieves rows from given query, then calls the passed
// ScanFunc for each row to store results directly into the Receiver (usually
// pointing to a slice of structures)
func (rq Request) GetIntoSlice(args ...interface{}) error {

	// Retrieve rows
	rows, err := rq.GetRows(args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Scan rows into receiver
	return scanToSlice(rows, rq.Arg)
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

// GetOneField makes a prepared query and returns the resulted row. This function
// can be used during and outside of transactions.
func (rq Request) GetOneField(args ...interface{}) error {

	// Retrieve row
	row, err := rq.GetOneRow(args...)
	if err != nil {
		return err
	}

	// Scan to ptr
	return scanToOnePtr(row, rq.Arg)
}

// GetFields retrieves the first row from given query, then calls the
// passed ScanFunc to store results directly into the Receiver (usually pointing
// to a slice)
func (rq Request) GetFields(args ...interface{}) error {

	// Retrieve row
	row, err := rq.GetOneRow(args...)
	if err != nil {
		return err
	}

	// Scan row into receiver
	return scanToSliceOfPtr(row, rq.Arg)
}

// GetIntoOneStruct retrieves the first row from given query, then calls the
// passed ScanFunc to store results directly into the Receiver (usually pointing
// to a structure)
func (rq Request) GetIntoOneStruct(args ...interface{}) error {

	// Retrieve row
	row, err := rq.GetOneRow(args...)
	if err != nil {
		return err
	}

	// Scan row into receiver
	return scanToOneStruct(row, rq.Arg)
}

// ExecQuery prepares a query which does not return a row, then calls the given
// ExecFunc with the passed Arg. This function can be used during and outside
// of transactions.
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
