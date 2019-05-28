package requests

import "database/sql"

// Check that rq.Arg is non-nil before making query
func (rq Request) getRowsWithArg(args ...interface{}) (*sql.Rows, error) {
	if rq.Arg == nil {
		return nil, ErrNoArg
	}
	return rq.GetRows(args...)
}

// Check that rq.Arg is non-nil before making query
func (rq Request) getOneRowWithArg(args ...interface{}) (*sql.Row, error) {
	if rq.Arg == nil {
		return nil, ErrNoArg
	}
	return rq.GetOneRow(args...)
}

// GetIntoStructs prepares and makes a query, retrieves Rows, and scan them into
// a pointer to a slice of structures. Should be used for queries selecting
// multiple fields from multiple rows
func (rq Request) GetIntoStructs(args ...interface{}) error {
	rows, err := rq.getRowsWithArg(args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return scanIntoStructs(rows, rq.Arg)
}

// GetIntoSlice prepares and makes a query, retrieves Rows, and scan them into
// a pointer to a slice of any type. Should be used for queries selecting one
// field from multiple rows
func (rq Request) GetIntoSlice(args ...interface{}) error {
	rows, err := rq.getRowsWithArg(args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return scanIntoSlice(rows, rq.Arg)
}

// GetOneField prepares and makes a query, retrieves a Row, and scan it into
// the passed pointer. Should be used for queries selecting one field from one
// row
func (rq Request) GetOneField(args ...interface{}) error {
	row, err := rq.getOneRowWithArg(args...)
	if err != nil {
		return err
	}
	return scanIntoOnePtr(row, rq.Arg)
}

// GetFields prepares and makes a query, retrieves a Row, and scan it into
// a slice of any pointers. Should be used for queries selecting multiple fields
// from a single row
func (rq Request) GetFields(args ...interface{}) error {
	row, err := rq.getOneRowWithArg(args...)
	if err != nil {
		return err
	}
	return scanIntoPtrs(row, rq.Arg)
}

// GetIntoOneStruct prepares and makes a query, retrieves a Row, and scan it
// into a pointer to one struct. Should be used for queries selecting multiple
// fields from a single row (if this query is unique, consider using GetFields
// instead)
func (rq Request) GetIntoOneStruct(args ...interface{}) error {
	row, err := rq.getOneRowWithArg(args...)
	if err != nil {
		return err
	}
	return scanToOneStruct(row, rq.Arg)
}
