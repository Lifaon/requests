package requests

// GetIntoStructs prepares and makes a query, retrieves Rows, and scan them into
// a slice of structures. Should be used for queries selecting multiple fields
// from multiple rows
func (rq Request) GetIntoStructs(args ...interface{}) error {

	// Retrieve rows
	rows, err := rq.GetRows(args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rq.Arg == nil {
		return ErrNoArg
	}

	// Scan rows into receiver
	return scanToSliceOfStruct(rows, rq.Arg)
}

// GetIntoSlice prepares and makes a query, retrieves Rows, and scan them into
// a slice of interface{}. Should be used for queries selecting one field from
// multiple rows
func (rq Request) GetIntoSlice(args ...interface{}) error {

	// Retrieve rows
	rows, err := rq.GetRows(args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rq.Arg == nil {
		return ErrNoArg
	}

	// Scan rows into receiver
	return scanToSlice(rows, rq.Arg)
}

// GetOneField prepares and makes a query, retrieves a Row, and scan it into
// a slice of interface{}. Should be used for queries selecting one field from
// one row
func (rq Request) GetOneField(args ...interface{}) error {

	// Retrieve row
	row, err := rq.GetOneRow(args...)
	if err != nil {
		return err
	}

	if rq.Arg == nil {
		return ErrNoArg
	}

	// Scan to ptr
	return scanToOnePtr(row, rq.Arg)
}

// GetFields prepares and makes a query, retrieves a Row, and scan it into
// a slice of pointers to interface{}. Should be used for queries selecting
// multiple fields from a single row
func (rq Request) GetFields(args ...interface{}) error {

	// Retrieve row
	row, err := rq.GetOneRow(args...)
	if err != nil {
		return err
	}

	if rq.Arg == nil {
		return ErrNoArg
	}

	// Scan row into receiver
	return scanToSliceOfPtr(row, rq.Arg)
}

// GetIntoOneStruct prepares and makes a query, retrieves a Row, and scan it
// into one struct. Should be used for queries selecting multiple fields
// from a single row (if this query is unique, use GetFields instead)
func (rq Request) GetIntoOneStruct(args ...interface{}) error {

	// Retrieve row
	row, err := rq.GetOneRow(args...)
	if err != nil {
		return err
	}

	if rq.Arg == nil {
		return ErrNoArg
	}

	// Scan row into receiver
	return scanToOneStruct(row, rq.Arg)
}
