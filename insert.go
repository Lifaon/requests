package requests

import (
	"database/sql"
	"fmt"
	"reflect"
)

// InsertStructs inserts a slice of structures into the given table. Can insert
// sub structures, if they have the tag `req:"include"`
func (rq Request) InsertStructs(slice interface{}) error {

	// Check that passed value is a slice
	elem := reflect.ValueOf(slice)
	if elem.Kind() != reflect.Slice {
		return fmt.Errorf("passed value should be a slice of structures, got: %s", elem.Type().String())
	}
	if elem.Len() == 0 {
		return fmt.Errorf("passed slice is empty")
	}
	// Check that passed slice stores structures
	st := elem.Type().Elem()
	if st.Kind() != reflect.Struct {
		return fmt.Errorf("passed slice should store structures, got: %s", st.String())
	}

	// Create insert query
	if err := rq.prepareInsert(st); err != nil {
		return err
	}
	// Prepare statement
	stmt, err := rq.PrepareStmt()
	if err != nil {
		return err
	}
	defer stmt.Close()
	// Insert each structure
	for i := 0; i < elem.Len(); i++ {
		if err := insertStruct(stmt, elem.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

// InsertOneStruct inserts one structure into the given table. Can insert sub
// structures, if they have the tag `req:"include"`
func (rq Request) InsertOneStruct(structure interface{}) error {
	// Check that passed value is a structure
	elem := reflect.ValueOf(structure)
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("passed value should be a structure, got: %s", elem.Type().String())
	}
	// Create insert query
	if err := rq.prepareInsert(elem.Type()); err != nil {
		return err
	}
	// Prepare statement
	stmt, err := rq.PrepareStmt()
	if err != nil {
		return err
	}
	defer stmt.Close()
	// Insert structure
	return insertStruct(stmt, elem)
}

// Write Query based on structure fields tags
func (rq *Request) prepareInsert(elem reflect.Type) error {
	// Retrieve column names
	columns := appendColumns([]string{}, elem)
	// Check that columns is not empty
	if len(columns) == 0 {
		return fmt.Errorf("passed structure has no field, or all fields are ignored")
	}
	// Write query
	rq.Statement = "INSERT INTO"
	rq.Set = "("
	rq.Condition = "VALUES ("
	for i, col := range columns {
		rq.Set += col
		rq.Condition += "?"
		if i < len(columns)-1 {
			rq.Set += ", "
			rq.Condition += ", "
		}
	}
	rq.Set += ")"
	rq.Condition += ")"
	return nil
}

// Create recursively a slice of all columns to insert from structure fields
// tags (to include sub structs)
func appendColumns(cols []string, elem reflect.Type) []string {
	// Run through each structure field
	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)
		// If the field is a structure, and marked to be included, recursively
		// call this function with this field
		if f.Type.Kind() == reflect.Struct && f.Tag.Get("req") == "include" {
			cols = appendColumns(cols, f.Type)
		} else {
			// Skip if no tag or ignored
			tag := f.Tag.Get("db")
			if tag == "" || tag == "-" {
				continue
			}
			// Store tag
			cols = append(cols, tag)
		}
	}
	return cols
}

// Execute insert query for one structure
func insertStruct(stmt *sql.Stmt, st reflect.Value) error {
	// Retrieve structure value
	var values []interface{}
	values = appendValues(values, st)
	// Make insert query
	_, err := stmt.Exec(values...)
	return err
}

// Create recursively a slice of all values to insert (to include sub structs)
func appendValues(values []interface{}, st reflect.Value) []interface{} {
	// Run through each structure field
	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)
		t := st.Type().Field(i)
		// If the field is a structure, and marked to be included, recursively
		// call this function with this field
		if t.Type.Kind() == reflect.Struct && t.Tag.Get("req") == "include" {
			values = appendValues(values, f)
		} else {
			// Skip if no tag or ignored
			tag := t.Tag.Get("db")
			if tag == "" || tag == "-" {
				continue
			}
			// Store value
			if f.Kind() != reflect.Ptr {
				values = append(values, f.Interface())
			} else if f.IsNil() {
				values = append(values, nil)
			} else {
				values = append(values, f.Elem().Interface())
			}
		}
	}
	return values
}
