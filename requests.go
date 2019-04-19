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
		Query      string      // full Query string
		Statement  string      // statement part of query
		Table      string      // targetted table of query
		Condition  string      // optionnal condition of query
		ScanFunc   scanFunc    // described below
		Receiver   interface{} // receiver for ScanFunc
		ExecFunc   execFunc    // described below
		Source     interface{} // source for ExecFunc
	}

	// Scan function which takes *sql.Row or *sql.Rows as a first argument, and
	// a pointer to where row.Scan will store results (a struct or a slice of structs)
	scanFunc func(row interface{}, receiver interface{}) error

	// Exec function which takes *sql.Stmt as a first argument, and executes the
	// prepared query with the given arguments
	execFunc func(stmt *sql.Stmt, source interface{}) error
)

// PrepareStmt prepares rq.Query via rq.SQLHandler. If rq.Query is empty, it
// will concatenate the query with rq.Statement, rq.Table, and rq.Condition
func (rq Request) PrepareStmt() (*sql.Stmt, error) {

	// If empty query, concatenate
	if rq.Query == "" {
		rq.Query = rq.Statement + " " + rq.Table + " " + rq.Condition
	}

	// Prepare statement
	switch r := rq.SQLHandler.(type) {
	case *sql.DB:
		return r.Prepare(rq.Query)
	case *sql.Tx:
		return r.Prepare(rq.Query)
	default:
		return nil, fmt.Errorf("wrong type passed to PrepareStmt: %T (expected *sql.DB or *sql.Tx)", r)
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

	// Check for scan function
	if rq.ScanFunc == nil {
		return fmt.Errorf("rq.ScanFunc is nil, consider using GetRows() or passing a valid function")
	}

	// Retrieve rows
	rows, err := rq.GetRows(args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Scan each row into receiver
	for rows.Next() {
		err = rq.ScanFunc(rows, rq.Receiver)
		if err != nil {
			return err
		}
	}

	return nil
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

	// Check for scan function
	if rq.ScanFunc == nil {
		return fmt.Errorf("rq.ScanFunc is nil, consider using GetOneRow() or passing a valid function")
	}

	// Retrieve row
	row, err := rq.GetOneRow(args...)
	if err != nil {
		return err
	}

	// Scan row into receiver
	return rq.ScanFunc(row, rq.Receiver)
}

// ScanRow scans from row (can be *sql.Row or *sql.Rows) into passed pointers
func ScanRow(row interface{}, pointers ...interface{}) error {

	// Scan row
	switch r := row.(type) {
	case *sql.Rows:
		return r.Scan(pointers...)
	case *sql.Row:
		return r.Scan(pointers...)
	default:
		return fmt.Errorf("wrong type passed to ScanRow: %T (expect *sql.Row or *sql.Rows)", r)
	}
}

// ExecQuery prepares a query which does not return a row, then calls the given
// ExecFunc with the passed Source. This function can be used during and outside
// of transactions.
func (rq Request) ExecQuery() error {

	// Prepare statement
	stmt, err := rq.PrepareStmt()
	if err != nil {
		return err
	}
	defer stmt.Close()

	// If an execFunc was given, execute it
	if rq.ExecFunc != nil {
		return rq.ExecFunc(stmt, rq.Source)
	}

	// Otherwise, execute statement
	_, err = stmt.Exec()
	return err
}
