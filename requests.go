package requests

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"
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

// scan and store values into pointed structure
func scanOneRow(row *sql.Row, ptr interface{}) error {

	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("passed value should be a pointer to structure, got: %T", v.Interface())
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("pointed value should be a structure, got: %T", elem.Interface())
	}

	// Scan into slice of interface
	values := make([]interface{}, elem.NumField())
	valuesptr := make([]interface{}, elem.NumField())
	for i := range values {
		valuesptr[i] = &(values[i])
	}
	row.Scan(valuesptr...)
	return storeToStruct(elem, values)
}

func scanRows(rows *sql.Rows, ptr interface{}) error {

	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("passed value should be a pointer to slice of structures, got: %T", v.Interface())
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Slice {
		return fmt.Errorf("pointed value should be a slice of structures, got: %T", elem.Interface())
	}
	if !elem.CanSet() {
		return fmt.Errorf("pointed value (of type %T) is not settable", elem.Kind())
	}

	if !rows.Next() {
		return nil
	}

	elem.Set(reflect.MakeSlice(elem.Type(), 1, 1))

	size := elem.Index(0).NumField()
	values := make([]interface{}, size)
	valuesptr := make([]interface{}, size)
	for i := range values {
		valuesptr[i] = &(values[i])
	}

	for i := 0; ; i++ {
		f := elem.Index(i)
		if !f.CanSet() {
			return fmt.Errorf("index #%d of created slice (of type %T) is not settable", i, elem.Kind())
		}
		err := rows.Scan(valuesptr...)
		if err != nil {
			return err
		}
		err = storeToStruct(f, values)
		if err != nil {
			return err
		}
		if rows.Next() {
			elem.Set(reflect.Append(elem, f))
		} else {
			break
		}
	}
	return nil
}

// store scanned values to structure
func storeToStruct(st reflect.Value, values []interface{}) error {
	for i, value := range values {
		// Retrieve structure field
		f := st.Field(i)
		if !f.CanSet() {
			return fmt.Errorf("structure field #%d of type %T is not settable", i, f.Interface())
		}
		// Set structure field
		switch val := value.(type) {
		case []byte:
			// struct field: []byte, *[]byte, string, *string, bool, or *bool
			s := string(val)
			b := false
			if len(val) != 0 && val[0] != 0 {
				b = true
			}
			switch f.Type() {
			case reflect.TypeOf(val): // []byte
				f.SetBytes(val)
			case reflect.TypeOf(&val): // *[]byte
				f.Set(reflect.ValueOf(&val))
			case reflect.TypeOf(s): // string
				f.SetString(s)
			case reflect.TypeOf(&s): // *string
				f.Set(reflect.ValueOf(&s))
			case reflect.TypeOf(b): // bool
				f.SetBool(b)
			case reflect.TypeOf(&b): // *bool
				f.Set(reflect.ValueOf(&b))
			default:
				return fmt.Errorf("structure field #%d doesn't have the right type (expected: []byte, *[]byte, string, *string, bool, or *bool, got: %T)", i, f.Interface())
			}
		case int64:
			// struct field: int64 or *int64
			switch f.Type() {
			case reflect.TypeOf(val): // int64
				f.SetInt(val)
			case reflect.TypeOf(&val): // *int64
				f.Set(reflect.ValueOf(&val))
			default:
				return fmt.Errorf("structure field #%d doesn't have the right type (expected: int64 or *int64, got: %T)", i, f.Interface())
			}
		case float64:
			// struct field: float64 or *float64
			switch f.Type() {
			case reflect.TypeOf(val): // float64
				f.SetFloat(val)
			case reflect.TypeOf(&val): // *float64
				f.Set(reflect.ValueOf(&val))
			default:
				return fmt.Errorf("structure field #%d doesn't have the right type (expected: float64 or *float64, got: %T)", i, f.Interface())
			}
		case time.Time:
			// struct field: time.Time or *time.Time
			switch f.Type() {
			case reflect.TypeOf(val): // time.Time
				f.Set(reflect.ValueOf(val))
			case reflect.TypeOf(&val): // *time.Time
				f.Set(reflect.ValueOf(&val))
			default:
				return fmt.Errorf("structure field #%d doesn't have the right type (expected: time.Time or *time.Time, got: %T)", i, f.Interface())
			}
		case nil:
			// struct field: any pointer
			if f.Kind() == reflect.Ptr {
				f.Set(reflect.Zero(f.Type()))
			} else {
				return fmt.Errorf("structure field #%d isn't a pointer when value can be <nil>, got: %T", i, f.Interface())
			}
		default:
			// unsupported type retrieved from *sql.Row(s).Scan()
			return fmt.Errorf("unsupported type retrieved from *sql.Row(s).Scan(): %T", val)
		}
	}
	return nil
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
